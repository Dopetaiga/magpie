package gateway

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

// Google's streamGenerateContent streams SSE only when it is asked with
// alt=sse; without it the reply is one JSON array of
// GenerateContentResponses, application/json, written element by
// element. magpie answered every stream with SSE, so a client calling
// the REST API directly and parsing the array got nothing it could read
// (#7297).
func TestGeminiArrayStreamWithoutAltSSE(t *testing.T) {
	f := dsUpstream(t)
	rec := httptest.NewRecorder()
	New().Handler().ServeHTTP(rec, httptest.NewRequest("POST", "/v1beta/models/ds/deepseek-v4.1-flash:streamGenerateContent", strings.NewReader(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`)))
	if rec.Code != 200 || f.calls != 1 {
		t.Fatalf("%d after %d calls: %s", rec.Code, f.calls, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type %q, want the array's JSON", ct)
	}
	var elems []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &elems); err != nil {
		t.Fatalf("not a JSON array: %v: %s", err, rec.Body.String())
	}
	if len(elems) == 0 {
		t.Fatal("an empty array")
	}
	cands, _ := elems[len(elems)-1]["candidates"].([]any)
	if len(cands) == 0 {
		t.Fatalf("the last element has no candidate: %v", elems[len(elems)-1])
	}
	last, _ := cands[0].(map[string]any)
	if last["finishReason"] == nil || last["finishReason"] == "" {
		t.Fatalf("the array's last element carries no finishReason: %v", last)
	}
}

// alt=sse stays the event stream it was.
func TestGeminiAltSSEStaysSSE(t *testing.T) {
	f := dsUpstream(t)
	rec := httptest.NewRecorder()
	New().Handler().ServeHTTP(rec, httptest.NewRequest("POST", "/v1beta/models/ds/deepseek-v4.1-flash:streamGenerateContent?alt=sse", strings.NewReader(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`)))
	if rec.Code != 200 || f.calls != 1 {
		t.Fatalf("%d after %d calls: %s", rec.Code, f.calls, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("Content-Type %q, want the event stream", ct)
	}
	if !strings.Contains(rec.Body.String(), "data: ") {
		t.Fatalf("not SSE: %s", rec.Body.String())
	}
}
