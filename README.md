
# Deprecated

Please use the [Go Beeline](https://github.com/honeycombio/beeline-go) instead, which has an updated `goji` middleware.

# Deprecated


a goji middleware to log request data to http://honeycomb.io

[![Build Status](https://travis-ci.org/honeycombio/goji-honey.svg?branch=master)](https://travis-ci.org/honeycombio/goji-honey)

If you're already using the fantastic http://goji.io/, it's trivial to
get request data flowing to a honeycomb dataset.  It adds all fields
of the http.Request object to the honeycomb event, as well as some
additions:

 - `GojiPattern`           useful as a normalized form of the request url
 - `$prefix_$variable`     all variables from the `GojiPattern` are part of the event, prefixed by the string you pass into `LogRequestToHoneycomb`.
 - `ResponseHttpStatus`    the http status code for this request
 - `ResponseContentLength` the amount of data written to the responseWriter
 - `ResponseTime_ms`       total response time in milliseconds, as measured in the middleware around inner handler.

For more information about using Honeycomb, check out our [docs](https://honeycomb.io/docs) and our [Go SDK](https://honeycomb.io/docs/connect/go/).

## Example

```go
package main

import (
        "fmt"
        "net/http"

        "goji.io"
        "goji.io/pat"
		
		"github.com/honeycombio/goji-honey"
)

func hello(w http.ResponseWriter, r *http.Request) {
        name := pat.Param(r, "name")
        fmt.Fprintf(w, "Hello, %s!", name)
}

func main() {
        mux := goji.NewMux()
		
		mux.Use(gojihoney.LogRequestToHoneycomb("gjv_"))
		
        mux.HandleFunc(pat.Get("/hello/:name"), hello)

        http.ListenAndServe("localhost:8000", mux)
}
```

Loading `http://localhost:8000/hello/Charity` will send an event to
honeycomb with `GojiPattern` = `/hello/:name`, and `gjv_name` =
`Charity`.

## Extending the event

The event is sent at the _end_ of request handling (in order to
include response time, http status code, and response length), and can
be extended by any inner handler during request handling, as in this
possible change to `hello` above:

```go
func hello(w http.ResponseWriter, r *http.Request) {
        name := pat.Param(ctx, "name")
		event := gojihoney.GetLibhoneyEvent(r.Context())
		
		before := time.Now()

		InsertHelloIntoDatabase(name)
		
		event.AddField("InsertTime_ms", time.Since(before).Seconds()*1000)
}
```

## Goji and the context package

Goji's API was recently updated to use Go 1.7's standard library
`context` package, which broke downstream unvendored packages like
ours.

`master` in goji-honey tracks Goji's `master`, and thus uses the new
`context`-based API.

Users of the old Goji API (found on Goji branch `net-context`) can
find a goji-honey implementation on a branch of the same name:
`net-context`.
