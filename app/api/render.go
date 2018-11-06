package api

import (
	"sync"
	"unsafe"

	"github.com/json-iterator/go"
	"github.com/labstack/echo"

	"github.com/Syncano/orion/pkg/settings"
)

var (
	jsonConfig     jsoniter.API
	jsonConfigOnce sync.Once
)

// RawMessage is universal []byte type that is registered for all available output encoders.
// Currently supports only JSON but it's easy to extend.
type RawMessage []byte

func init() {
	jsoniter.RegisterTypeEncoderFunc("api.RawMessage", func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
		stream.WriteRaw(string(*((*[]byte)(ptr))))
	}, func(unsafe.Pointer) bool {
		return false
	})
}

// Render serializes and outputs object to http writer depending on content negotiation..
// Currently supports only JSON but it's easy to extend.
func Render(e echo.Context, code int, obj interface{}) error {
	bytes, err := Marshal(e, obj)
	if err != nil {
		return err
	}
	return e.JSONBlob(code, bytes)
}

// Marshal serializes object depending on content negotiation.
// Currently supports only JSON but it's easy to extend.
func Marshal(c echo.Context, obj interface{}) ([]byte, error) {
	jsonConfigOnce.Do(func() {
		jsonConfig = jsoniter.Config{
			ObjectFieldMustBeSimpleString: true,                  // do not unescape object field
			SortMapKeys:                   settings.Common.Debug, // sort map keys if debug mode is on
		}.Froze()
	})
	return jsonConfig.Marshal(obj)
}
