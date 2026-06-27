package handler

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/snakesneaks/webhook-relay/cfg"
	"github.com/snakesneaks/webhook-relay/service"
)

type handler struct {
	cfg           *cfg.Config
	renderService service.RenderService
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	route := h.cfg.FindRoute(req.URL.Path)
	if route == nil {
		http.NotFound(rw, req)
		return
	}

	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, "failed to read body", http.StatusBadRequest)
		log.Println(err)
		return
	}
	defer req.Body.Close()

	target, outHeaders, outBody, err := h.renderService.Render(route, req, rawBody)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		log.Println(err)
		return
	}

	forwardReq, err := http.NewRequest(
		http.MethodPost,
		target,
		bytes.NewReader(outBody),
	)
	if err != nil {
		http.Error(rw, "failed to build request", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	for k, v := range outHeaders {
		forwardReq.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(forwardReq)
	if err != nil {
		http.Error(rw, "failed to forward webhook", http.StatusBadGateway)
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(rw, "failed to read upstream response", http.StatusBadGateway)
		log.Println(err)
		return
	}

	for k, vals := range resp.Header {
		for _, v := range vals {
			rw.Header().Add(k, v)
		}
	}

	rw.WriteHeader(resp.StatusCode)
	_, _ = rw.Write(respBody)
}

func NewRelayHandler(cfg *cfg.Config, r service.RenderService) http.Handler {
	return &handler{
		cfg:           cfg,
		renderService: r,
	}
}
