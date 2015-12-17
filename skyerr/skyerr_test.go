package skyerr

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"fmt"
)

func TestNewError(t *testing.T) {
	Convey("An Error", t, func() {
		err := NewError(10000, "some message")

		Convey("returns code correctly", func() {
			So(err.Code(), ShouldEqual, 10000)
		})

		Convey("returns message correctly", func() {
			So(err.Message(), ShouldEqual, "some message")
		})

		Convey("Error()s in format {code}: {message}", func() {
			So(err.Error(), ShouldEqual, "UnexpectedError: some message")
		})

		Convey("has format {code}: {message} when being written", func() {
			So(fmt.Sprintf("%v", err), ShouldEqual, "UnexpectedError: some message")
		})
	})
}

func TestNewErrorf(t *testing.T) {
	Convey("NewErrorf", t, func() {
		err := NewErrorf(2, "obj1: %v, obj2: %v", "string", 0)

		Convey("creates err with correct code", func() {
			So(err.Code(), ShouldEqual, 2)
		})

		Convey("creates err with correct message", func() {
			So(err.Message(), ShouldEqual, "obj1: string, obj2: 0")
		})
	})
}
