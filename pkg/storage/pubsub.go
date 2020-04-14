package storage

import (
	"sync"

	"github.com/go-redis/redis/v7"
)

type PubSub struct {
	mu       sync.RWMutex
	initOnce sync.Once
	cli      *redis.Client

	subs   map[string][]chan<- string
	pubsub *redis.PubSub
}

func NewPubSub(cli *redis.Client) *PubSub {
	return &PubSub{
		cli:  cli,
		subs: make(map[string][]chan<- string),
	}
}

func (p *PubSub) Subscribe(name string, ch chan<- string) error {
	p.initOnce.Do(func() {
		p.pubsub = p.cli.Subscribe()
		go p.process()
	})

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.subs[name]; !ok {
		if err := p.pubsub.Subscribe(name); err != nil {
			return err
		}
	}

	p.subs[name] = append(p.subs[name], ch)

	return nil
}

func (p *PubSub) Unsubscribe(name string) error {
	return p.pubsub.Unsubscribe(name)
}

func (p *PubSub) process() {
	ch := p.pubsub.Channel()

	for msg := range ch {
		p.mu.RLock()
		out := p.subs[msg.Channel]
		p.mu.RUnlock()

		for _, o := range out {
			select {
			case o <- msg.Payload:
			default:
			}
		}
	}
}
