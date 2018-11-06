package celery

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"

	"github.com/Syncano/orion/pkg/celery/mocks"
)

var err = errors.New("some error")

type obj int

func (a obj) MarshalJSON() ([]byte, error) {
	return nil, err
}

func TestCelery(t *testing.T) {
	amqpCh := new(mocks.AMQPChannel)
	queue := "queue"
	Init(amqpCh)

	Convey("NewTask works with nil args and kwargs", t, func() {
		task := NewTask("sometask", queue, nil, nil)
		So(len(task.Args), ShouldEqual, 0)
		So(len(task.Kwargs), ShouldEqual, 0)

		Convey("Publish publishes task to amqp", func() {
			amqpCh.On("Publish", mock.Anything, queue, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			e := task.Publish()
			So(e, ShouldBeNil)
			amqpCh.AssertExpectations(t)
		})
		Convey("Publish propagates json marshal error", func() {
			task.Args = []interface{}{obj(0)}
			e := task.Publish()
			So(e, ShouldNotBeNil)
			So(e.Error(), ShouldContainSubstring, err.Error())
			amqpCh.AssertExpectations(t)
		})
	})
}
