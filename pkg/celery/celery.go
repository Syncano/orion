package celery

import (
	"time"

	json "github.com/json-iterator/go"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

var (
	amqpCh AMQPChannel
)

const (
	exchange = ""
)

// Task describes celery task definition.
type Task struct {
	ch     AMQPChannel
	Task   string                 `json:"task"`
	Queue  string                 `json:"-"`
	ID     string                 `json:"id"`
	Args   []interface{}          `json:"args"`
	Kwargs map[string]interface{} `json:"kwargs"`
}

type Celery struct {
	ch AMQPChannel
}

// Init sets up celery tasks.
func New(ch AMQPChannel) *Celery {
	return &Celery{
		ch: ch,
	}
}

// NewTask returns a new task object.
func NewTask(task, queue string, args []interface{}, kwargs map[string]interface{}) *Task {
	if args == nil {
		args = make([]interface{}, 0)
	}

	if kwargs == nil {
		kwargs = make(map[string]interface{})
	}

	return &Task{
		ch:     amqpCh,
		Task:   task,
		Queue:  queue,
		ID:     uuid.NewV4().String(),
		Args:   args,
		Kwargs: kwargs,
	}
}

// Publish sends task to channel.
func (t *Task) Publish(c *Celery) error {
	body, err := json.Marshal(t)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{
		DeliveryMode:    amqp.Persistent,
		Timestamp:       time.Now(),
		ContentType:     "application/json",
		ContentEncoding: "utf-8",
		Body:            body,
	}

	return c.ch.Publish(exchange, t.Queue, false, false, msg)
}
