package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/myselfBZ/chatrix/internal/auth"
	"github.com/myselfBZ/chatrix/internal/store"
)

func newTestServer() *Server {
	s := &Server{}

	s.store = store.NewMockStore()
    s.auth = &auth.AuthMock{}

    return s
}



func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d", expected, actual)
	}
}

