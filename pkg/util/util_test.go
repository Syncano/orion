package util

import (
	"errors"
	"io"
	"testing"
	"time"
	"unicode/utf8"

	json "github.com/json-iterator/go"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestGenerateRandomString(t *testing.T) {
	Convey("Given some length, generate returns random strings", t, func() {
		val := GenerateRandomString(10)
		So(val, ShouldNotEqual, GenerateRandomString(10))
		So(len(val), ShouldEqual, 10)
	})
}

func TestGenerateKey(t *testing.T) {
	Convey("Generate key returns random keys of constant length", t, func() {
		val1 := GenerateKey()
		val2 := GenerateKey()
		So(val1, ShouldNotEqual, val2)
		So(len(val1), ShouldEqual, len(val2))
		So(len(val1), ShouldEqual, 40)
	})
}

type MockObject struct {
	mock.Mock
}

func (m *MockObject) f() error {
	return m.Called().Error(0)
}

func (m *MockObject) f2() (bool, error) {
	ret := m.Called()
	return ret.Bool(0), ret.Error(1)
}

func TestRetry(t *testing.T) {
	err := errors.New("some error")

	Convey("Given some error func", t, func() {
		mo := MockObject{}

		Convey("Retry respects sleep time", func() {
			sleep := 2 * time.Millisecond
			mo.On("f").Return(err)
			t := time.Now()
			Retry(2, sleep, mo.f)
			So(time.Since(t), ShouldBeGreaterThan, sleep)
		})
		Convey("Retry returns error after X attempts", func() {
			mo.On("f").Return(err).Times(10)
			e := Retry(10, 0, mo.f)
			So(e, ShouldEqual, err)
		})
		Convey("Retry returns nil if it succeeds", func() {
			mo.On("f").Return(io.EOF).Once()
			mo.On("f").Return(nil).Once()
			e := Retry(10, 0, mo.f)
			So(e, ShouldBeNil)
		})

		Convey("RetryWithCritical returns error after X attempts", func() {
			mo.On("f2").Return(false, err).Times(10)
			e := RetryWithCritical(10, 0, mo.f2)
			So(e, ShouldEqual, err)
		})
		Convey("RetryWithCritical returns error on critical", func() {
			mo.On("f2").Return(true, err).Once()
			e := RetryWithCritical(10, 0, mo.f2)
			So(e, ShouldEqual, err)
		})
		mo.AssertExpectations(t)
	})
}

func TestUtil(t *testing.T) {
	Convey("Must panics for non-empty error", t, func() {
		So(func() { Must(io.EOF) }, ShouldPanic)
	})

	Convey("Hash creates unique and consistent hash", t, func() {
		So(Hash("a"), ShouldEqual, Hash("a"))
		So(Hash("a"), ShouldNotEqual, Hash("b"))
	})
}

func TestToQuoteJSON(t *testing.T) {
	Convey("Given []byte of utf8 string", t, func() {
		str := []byte("ąęść\"ABC\"\a\b\f\t\r\n\v\u2318\U0010FFFF百度")
		buf := make([]byte, 4)
		n := utf8.EncodeRune(buf, rune(0x1FFFFFFF))
		str = append(str, buf[:n]...)

		Convey("ToQuoteJSON produces valid ascii JSON", func() {
			data := ToQuoteJSON(str)
			_, err := json.Marshal(json.RawMessage(data))
			So(err, ShouldBeNil)
			So(string(ToQuoteJSON(str)), ShouldEqual, `"\u0105\u0119\u015b\u0107\"ABC\"\u0007\b\f\t\r\n\u000b\u2318\udbff\udfff\u767e\u5ea6\ufffd"`)
		})
	})
}
