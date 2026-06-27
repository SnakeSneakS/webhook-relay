package webhookrelay

import (
	"log"
	"net/http"

	"github.com/snakesneaks/webhook-relay/cfg"
	"github.com/snakesneaks/webhook-relay/handler"
)

func main() {
	cfg, err := cfg.Load()
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()

	// health endpoints
	mux.Handle("/healthz", handler.NewHealthHandler())

	// relay handler
	mux.Handle("/", handler.NewRelayHandler(cfg))

	if err := http.ListenAndServe(cfg.Server.Addr, mux); err != nil {
		log.Fatal(err)
	}
}
