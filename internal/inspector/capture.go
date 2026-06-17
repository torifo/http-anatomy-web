// Package inspector turns an HTTP request plus the primary response
// fragment into a model.Exchange for the HTTP inspector pane.
package inspector

import (
	"net/http"

	"http-anatomy/internal/model"
)

// shownHeaders lists the request headers surfaced in the inspector, in
// display order. HX-* headers are the teaching highlight of HTMX.
var shownHeaders = []string{
	"Host",
	"HX-Request",
	"HX-Target",
	"HX-Trigger",
	"Content-Type",
}

// respContentType is the content type every fragment response uses.
const respContentType = "text/html; charset=utf-8"

// shownRespHeaders lists the response headers surfaced in the inspector, in
// display order. These HTMX response headers drive client-side behavior
// (events, redirects, swap overrides) and are a teaching highlight.
var shownRespHeaders = []string{
	"HX-Trigger",
	"HX-Reswap",
	"HX-Retarget",
	"HX-Redirect",
}

// BuildExchange captures one request/response for the inspector. body is
// the primary swap fragment only; the inspector's own OOB block must never
// be passed in, so the inspector never renders itself. respHeaders is the
// response header set already written by the handler.
func BuildExchange(r *http.Request, body string, status int, respHeaders http.Header) model.Exchange {
	var reqHeaders []model.Header
	for _, name := range shownHeaders {
		v := r.Header.Get(name)
		if name == "Host" && v == "" {
			v = r.Host // Host is not stored in r.Header.
		}
		if v != "" {
			reqHeaders = append(reqHeaders, model.Header{Name: name, Value: v})
		}
	}
	var resHeaders []model.Header
	for _, name := range shownRespHeaders {
		if v := respHeaders.Get(name); v != "" {
			resHeaders = append(resHeaders, model.Header{Name: name, Value: v})
		}
	}
	return model.Exchange{
		Method:      r.Method,
		Path:        r.URL.Path,
		Proto:       r.Proto,
		ReqHeaders:  reqHeaders,
		Status:      status,
		StatusText:  http.StatusText(status),
		RespCType:   respContentType,
		RespHeaders: resHeaders,
		Body:        body,
	}
}
