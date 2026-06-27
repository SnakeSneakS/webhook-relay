package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/snakesneaks/webhook-relay/cfg"
)

type handler struct {
	cfg *cfg.Config
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
		return
	}
	defer req.Body.Close()

	outHeaders, outBody, err := renderRoute(route, req, rawBody)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	forwardReq, err := http.NewRequest(
		http.MethodPost,
		route.Target,
		bytes.NewReader(outBody),
	)
	if err != nil {
		http.Error(rw, "failed to build request", http.StatusInternalServerError)
		return
	}

	for k, v := range outHeaders {
		forwardReq.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(forwardReq)
	if err != nil {
		http.Error(rw, "failed to forward webhook", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(rw, "failed to read upstream response", http.StatusBadGateway)
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

func NewRelayHandler(cfg *cfg.Config) http.Handler {
	return &handler{
		cfg: cfg,
	}
}

func renderRoute(
	route *cfg.Route,
	req *http.Request,
	rawBody []byte,
) (map[string]string, []byte, error) {
	var bodyData map[string]interface{}

	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &bodyData); err != nil {
			return nil, nil, fmt.Errorf("invalid json body: %w", err)
		}
	} else {
		bodyData = map[string]interface{}{}
	}

	funcs := template.FuncMap{
		"header": func(key string) string {
			return req.Header.Get(key)
		},
		"body": func(path string) interface{} {
			return resolvePath(bodyData, path)
		},
		"env": func(key string, fallback ...string) string {
			v := os.Getenv(key)
			if v != "" {
				return v
			}

			if len(fallback) > 0 {
				return fallback[0]
			}

			return ""
		},
	}

	render := func(tmpl string) (string, error) {
		t, err := template.New("route").Funcs(funcs).Parse(tmpl)
		if err != nil {
			return "", err
		}

		var buf bytes.Buffer
		if err := t.Execute(&buf, nil); err != nil {
			return "", err
		}

		return buf.String(), nil
	}

	outHeaders := make(map[string]string)

	for key, tmpl := range route.Headers {
		rendered, err := render(tmpl)
		if err != nil {
			return nil, nil, fmt.Errorf("render header %s: %w", key, err)
		}
		outHeaders[key] = rendered
	}

	renderedBody, err := render(route.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("render body: %w", err)
	}

	return outHeaders, []byte(renderedBody), nil
}

func resolvePath(data map[string]interface{}, path string) interface{} {
	current := interface{}(data)

	for _, part := range strings.Split(path, ".") {
		obj, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}

		next, exists := obj[part]
		if !exists {
			return nil
		}

		current = next
	}

	return current
}
