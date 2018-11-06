package amqp

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/streadway/amqp"
	"go.uber.org/zap"

	"github.com/Syncano/orion/pkg/log"
)

const (
	amqpRetrySleep    = 250 * time.Millisecond
	amqpMaxRetrySleep = 2 * time.Second
)

// Channel is a wrapper for amqp channel supporting automatic reconnect.
type Channel struct {
	mu               sync.Mutex
	ch               *amqp.Channel
	running          uint32
	registeredQueues map[string]struct{}
}

func (ac *Channel) connect(url string) error {
	ac.ch = nil
	connection, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	ch, err := connection.Channel()
	if err != nil {
		connection.Close() // nolint: errcheck
		return err
	}
	if err = ch.ExchangeDeclare(
		"default", // name
		"direct",  // type
		true,      // durable
		false,     // auto-deleted
		false,     // internal
		false,     // noWait
		nil,       // arguments
	); err != nil {
		ch.Close()         // nolint: errcheck
		connection.Close() // nolint: errcheck
		return err
	}
	ac.ch = ch
	ac.registeredQueues = make(map[string]struct{})
	return nil
}

// Init creates amqp channel with specified url and retry mechanism.
func (ac *Channel) Init(url string) error {
	logger := log.Logger()

	ac.mu.Lock()
	err := ac.connect(url)
	ac.mu.Unlock()

	if err != nil {
		return err
	}

	ac.setRunning(true)

	// Start connection monitor.
	go func() {
		for {
			amqpCloseCh := make(chan *amqp.Error)
			ac.ch.NotifyClose(amqpCloseCh)
			e := <-amqpCloseCh

			if e != nil {
				logger.With(zap.Error(e)).Warn("Lost AMQP connection")
				amqpSleep := amqpRetrySleep
				ac.mu.Lock()

				for {
					if ac.IsRunning() {
						if e := ac.connect(url); e != nil {
							logger.With(zap.Error(e)).Error("Cannot connect to AMQP, retrying")
							time.Sleep(amqpSleep)
							if amqpSleep < amqpMaxRetrySleep {
								amqpSleep += amqpRetrySleep
							}
							continue
						}
						logger.Info("Reconnected to AMQP")
					}
					break
				}
				ac.mu.Unlock()
			} else {
				logger.Info("Lost AMQP connection (graceful stop)")
				break
			}
		}
	}()

	return nil
}

// IsRunning returns true if channel is setup and running.
func (ac *Channel) IsRunning() bool {
	return (atomic.LoadUint32(&ac.running) == 1)
}

func (ac *Channel) setRunning(running bool) {
	if running {
		atomic.StoreUint32(&ac.running, 1)
	} else {
		atomic.StoreUint32(&ac.running, 0)
	}
}

// Publish sends a Publishing from the client to an exchange on the server.
func (ac *Channel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	if _, ok := ac.registeredQueues[key]; !ok {
		if _, err := ac.ch.QueueDeclare(
			key,   // name
			true,  // durable
			false, // autoDelete
			false, // exclusive
			false, // noWait
			nil,   // args
		); err != nil {
			return err
		}
		ac.registeredQueues[key] = struct{}{}
	}

	return ac.ch.Publish(exchange, key, mandatory, immediate, msg)
}

// Shutdown stops gracefully Channel.
func (ac *Channel) Shutdown() {
	ac.setRunning(false)
	ac.mu.Lock()
	if ac.ch != nil {
		ac.ch.Close()
	}
	ac.mu.Unlock()
}
