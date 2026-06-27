package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/snakesneaks/webhook-relay/cfg"
)

type renderService struct {
}

type RenderService interface {
	Render(
		route *cfg.Route,
		req *http.Request,
		rawBody []byte,
	) (target string, headers map[string]string, body []byte, err error)
}

func NewRenderService() RenderService {
	return &renderService{}
}

func (r *renderService) Render(
	route *cfg.Route,
	req *http.Request,
	rawBody []byte,
) (target string, headers map[string]string, body []byte, err error) {
	var bodyData map[string]interface{}

	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &bodyData); err != nil {
			return "", nil, nil, fmt.Errorf("invalid json body: %w", err)
		}
	} else {
		bodyData = map[string]interface{}{}
	}

	funcs := template.FuncMap{
		"header": func(key string) string {
			return req.Header.Get(key)
		},
		"body": func(paths ...string) interface{} {
			if len(paths) == 0 {
				return bodyData
			}
			return r.resolvePath(bodyData, paths[0])
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
		"toJson": func(v interface{}) string {
			return r.toJSON(v, false)
		},
		"toPrettyJson": func(v interface{}) string {
			return r.toJSON(v, true)
		},
		"escape": func(s string) string {
			b, err := json.Marshal(s)
			if err != nil {
				return ""
			}

			escaped := string(b)
			return escaped[1 : len(escaped)-1]
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
			return "", nil, nil, fmt.Errorf("render header %s: %w", key, err)
		}
		outHeaders[key] = rendered
	}

	renderedBody, err := render(route.Body)
	if err != nil {
		return "", nil, nil, fmt.Errorf("render body: %w", err)
	}

	renderedTarget, err := render(route.Target)
	if err != nil {
		return "", nil, nil, fmt.Errorf("render target: %w", err)
	}

	return renderedTarget, outHeaders, []byte(renderedBody), nil
}

func (r *renderService) toJSON(v interface{}, pretty bool) string {
	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)

	if pretty {
		enc.SetIndent("", "  ")
	}

	if err := enc.Encode(v); err != nil {
		return ""
	}

	return strings.TrimSpace(buf.String())
}

func (r *renderService) resolvePath(data map[string]interface{}, path string) interface{} {
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
