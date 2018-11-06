package celery

import "github.com/streadway/amqp"

// AMQPChannel defines amqp channel methods we are using.
//go:generate mockery -name AMQPChannel
type AMQPChannel interface {
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
}

// Assert that amqp channel is compatible with our interface.
var _ AMQPChannel = &amqp.Channel{}
