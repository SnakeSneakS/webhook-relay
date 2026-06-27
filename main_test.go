package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/snakesneaks/webhook-relay/cfg"
	relayhandler "github.com/snakesneaks/webhook-relay/handler"
	"github.com/snakesneaks/webhook-relay/service"
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
	renderService := service.NewRenderService()

	h := relayhandler.NewRelayHandler(config, renderService)

	req := httptest.NewRequest(http.MethodPost, "/unknown", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
