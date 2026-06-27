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
	if err := http.ListenAndServe(cfg.Server.Addr, handler.NewHandler(cfg)); err != nil {
		log.Fatal(err)
	}
}
