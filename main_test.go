package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/snakesneaks/webhook-relay/cfg"
	relayhandler "github.com/snakesneaks/webhook-relay/handler"
)

func TestHandlerCreation(t *testing.T) {
	config := &cfg.Config{
		Server: cfg.ServerConfig{
			Addr: ":8080",
		},
		App: cfg.AppConfig{
			Routes: []cfg.Route{
				{
					Path:   "/test",
					Target: "http://example.com",
				},
			},
		},
	}

	h := relayhandler.NewRelayHandler(config)

	req := httptest.NewRequest(http.MethodPost, "/unknown", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
