package version

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParsing(t *testing.T) {
	Convey("Given some build time string", t, func() {
		Convey("of proper time format, parsing succeeds", func() {
			t := mustParseBuildtime("2010-02-01T11:01")
			So(t, ShouldResemble, time.Date(2010, 02, 01, 11, 01, 0, 0, time.UTC))
		})
		Convey("of improper value, parsing panics", func() {
			So(func() { mustParseBuildtime("bla") }, ShouldPanic)
		})
	})
}
