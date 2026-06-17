package inspector

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildExchangeSelectsHeaders(t *testing.T) {
	r := httptest.NewRequest("DELETE", "/api/todos/42", nil)
	r.Host = "http-anatomy.dev"
	r.Header.Set("HX-Request", "true")
	r.Header.Set("HX-Target", "todo-42")
	r.Header.Set("Authorization", "secret") // not in shownHeaders
	r.Header.Set("Cookie", "ha_session=x")  // not in shownHeaders

	respHeaders := http.Header{}
	respHeaders.Set("HX-Trigger", "showToast")
	ex := BuildExchange(r, "<li>fragment</li>", 200, respHeaders)

	if len(ex.RespHeaders) != 1 || ex.RespHeaders[0].Name != "HX-Trigger" || ex.RespHeaders[0].Value != "showToast" {
		t.Fatalf("HX-Trigger response header not captured: %+v", ex.RespHeaders)
	}
	if ex.Method != "DELETE" || ex.Path != "/api/todos/42" {
		t.Fatalf("method/path wrong: %s %s", ex.Method, ex.Path)
	}
	if ex.Status != 200 || ex.StatusText != "OK" {
		t.Fatalf("status text wrong: %d %q", ex.Status, ex.StatusText)
	}
	names := map[string]string{}
	for _, h := range ex.ReqHeaders {
		names[h.Name] = h.Value
	}
	if names["Host"] != "http-anatomy.dev" {
		t.Fatalf("Host not captured: %q", names["Host"])
	}
	if names["HX-Target"] != "todo-42" {
		t.Fatalf("HX-Target not captured: %q", names["HX-Target"])
	}
	if _, ok := names["Authorization"]; ok {
		t.Fatal("Authorization should not be surfaced")
	}
	if _, ok := names["Cookie"]; ok {
		t.Fatal("Cookie should not be surfaced")
	}
}

func TestBuildExchangeBodyIsPrimaryOnly(t *testing.T) {
	r := httptest.NewRequest("POST", "/api/todos", nil)
	body := "<li id=\"todo-1\">buy milk</li>"
	ex := BuildExchange(r, body, 201, http.Header{})
	if ex.Body != body {
		t.Fatalf("body should be the primary fragment verbatim, got %q", ex.Body)
	}
	if strings.Contains(ex.Body, "hx-swap-oob") {
		t.Fatal("inspector body must not contain its own OOB block")
	}
	if ex.StatusText != "Created" {
		t.Fatalf("status text wrong: %q", ex.StatusText)
	}
}
