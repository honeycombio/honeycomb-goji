package gojihoney_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"goji.io"
	"goji.io/pat"

	"github.com/honeycombio/goji-honey"
	"github.com/honeycombio/libhoney-go"
)

func testEquals(t testing.TB, actual, expected interface{}, msg ...string) {
	if !reflect.DeepEqual(actual, expected) {
		testCommonErr(t, actual, expected, msg)
	}
}

func testCommonErr(t testing.TB, actual, expected interface{}, msg []string) {
	message := strings.Join(msg, ", ")
	_, file, line, _ := runtime.Caller(2)

	t.Errorf(
		"%s:%d: %s -- actual(%T): %v, expected(%T): %v",
		filepath.Base(file),
		line,
		message,
		testDeref(actual),
		testDeref(actual),
		testDeref(expected),
		testDeref(expected),
	)
}

func testDeref(v interface{}) interface{} {
	switch t := v.(type) {
	case *string:
		return fmt.Sprintf("*(%v)", *t)
	case *int64:
		return fmt.Sprintf("*(%v)", *t)
	case *float64:
		return fmt.Sprintf("*(%v)", *t)
	case *bool:
		return fmt.Sprintf("*(%v)", *t)
	default:
		return v
	}
}

type testTransport struct {
	invoked bool
	event   map[string]interface{}
}

func (tr *testTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	tr.invoked = true
	b, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(b, &tr.event)
	return &http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil
}

func hello(w http.ResponseWriter, r *http.Request) {
	name := pat.Param(r, "name")
	fmt.Fprintf(w, "Hello, %s!", name)
}

func TestMiddleware(t *testing.T) {
	tr := &testTransport{}

	libhoney.Init(libhoney.Config{
		WriteKey:   "aoeu",
		Dataset:    "oeui",
		SampleRate: 1,
		APIHost:    "http://localhost:8081/",
		Transport:  tr,
	})

	mux := goji.NewMux()
	mux.Use(gojihoney.LogRequestToHoneycomb("gjv_"))
	mux.HandleFunc(pat.Get("/hello/:name"), hello)

	r, _ := http.NewRequest("GET", "/hello/boris", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	libhoney.Close()
	testEquals(t, tr.invoked, true)

	testEquals(t, tr.event["gjv_name"], "boris")
	testEquals(t, tr.event["GojiPattern"], "/hello/:name")
	testEquals(t, tr.event["ResponseHttpStatus"], float64(200))
}
