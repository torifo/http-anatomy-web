package model

// Header is a single HTTP header surfaced in the inspector.
type Header struct {
	Name  string
	Value string
}

// Exchange is one request/response captured for the HTTP inspector.
// Body holds the primary swap fragment only (never the inspector's own
// out-of-band block), so the inspector never renders itself recursively.
type Exchange struct {
	Method     string   // "DELETE"
	Path       string   // "/api/todos/42"
	Proto      string   // "HTTP/1.1"
	ReqHeaders []Header // Host, HX-Request, HX-Target, HX-Trigger, Content-Type
	Status     int      // 200
	StatusText string   // "OK"
	RespCType  string   // "text/html; charset=utf-8"
	Body       string   // primary fragment, shown escaped
}
