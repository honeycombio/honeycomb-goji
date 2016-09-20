package gojihoney

import (
	"net/http"
	"time"

	"goji.io"
	"goji.io/middleware"
	"goji.io/pat"
	"goji.io/pattern"
	"golang.org/x/net/context"

	libhoney "github.com/honeycombio/libhoney-go"
)

const (
	libhoneyEventContextKey = "libhoneyEvent"
)

func GetLibhoneyEvent(ctx context.Context) *libhoney.Event {
	if event, ok := ctx.Value(libhoneyEventContextKey).(*libhoney.Event); ok {
		return event
	}
	return nil
}

// Middleware: log http.Requests and response HTTP status/content-length/time
// to Hound.
func LogRequestToHoneycomb(varPrefix string) func(goji.Handler) goji.Handler {
	return func(handler goji.Handler) goji.Handler {
		return goji.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			before := time.Now()

			event := libhoney.NewEvent()
			event.Add(r)

			gojiPattern := middleware.Pattern(ctx)
			if gojiPattern != nil {
				// log our pattern
				event.AddField("GojiPattern", gojiPattern.(*pat.Pattern).String())
			}

			// and the variables
			if variables, ok := ctx.Value(pattern.AllVariables).(map[pattern.Variable]interface{}); ok {
				for k, v := range variables {
					event.AddField(varPrefix+string(k), v.(string))
				}
			}

			responseWriter := newResponseWriterProxy(w)
			handler.ServeHTTPC(context.WithValue(ctx, libhoneyEventContextKey, event), responseWriter, r)

			event.AddField("ResponseHttpStatus", responseWriter.Status())
			event.AddField("ResponseContentLength", responseWriter.Length())
			event.AddField("ResponseTime_ms", time.Since(before).Seconds()*1000)

			event.Send()
		})
	}
}

type responseWriterProxy struct {
	http.ResponseWriter
	statusCode int
	length     int
}

func newResponseWriterProxy(inner http.ResponseWriter) *responseWriterProxy {
	return &responseWriterProxy{inner, 0, 0}
}
func (rw *responseWriterProxy) Status() int {
	return rw.statusCode
}
func (rw *responseWriterProxy) Length() int {
	return rw.length
}
func (rw *responseWriterProxy) Write(bytes []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = 200
	}
	rv, err := rw.ResponseWriter.Write(bytes)
	rw.length += rv
	return rv, err
}
func (rw *responseWriterProxy) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
