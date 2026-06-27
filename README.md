# webhook-relay

A lightweight webhook relay written in Go.

`webhook-relay` receives incoming webhooks, transforms headers and payloads using Go templates, and forwards them to another webhook endpoint.

## Features

- Lightweight single-binary server
- Docker-friendly
- YAML-based configuration
- Transform headers and body with Go templates
- Relay webhooks to arbitrary HTTP endpoints
- Dynamic payload transformation

---

## Installation

### Build from source

```bash
go build -o webhook-relay ./cmd/webhook-relay
```

### Run

```bash
ADDR=:8080 CONFIG_PATH=./config.yaml ./webhook-relay
```

---

## Configuration

Configuration is split into:

- Environment variables (server settings)
- YAML file (relay routes)

### Environment variables

| Variable | Default | Description |
|---|---|---|
| `ADDR` | `:8080` | HTTP listen address |
| `CONFIG_PATH` | `config.yaml` | Path to YAML config |

---

## Example `config.yaml`

```yaml
routes:
  - path: /github
    target: https://example.com/webhook

    headers:
      Authorization: "Bearer {{ header `X-Token` }}"
      X-Event: "{{ header `X-GitHub-Event` }}"

    body: |
      {
        "repository": "{{ body `repository.name` }}",
        "sender": "{{ body `sender.login` }}",
        "action": "{{ body `action` }}"
      }

  - path: /slack
    target: https://example.com/slack

    headers:
      Content-Type: application/json

    body: |
      {
        "text": "{{ body `message.text` }}"
      }
```

---

## Route Configuration

Each route contains:

| Field | Required | Description |
|---|---|---|
| `path` | yes | Incoming webhook path |
| `target` | yes | Destination webhook URL |
| `headers` | no | Outgoing headers |
| `body` | no | Outgoing request body template |

---

## Template Functions

### `header`

Reads incoming request headers.

```gotemplate
{{ header "Authorization" }}
```

Example:

```yaml
headers:
  Authorization: "{{ header `Authorization` }}"
```

---

### `body`

Reads values from JSON payload using dot notation.

Incoming payload:

```json
{
  "repository": {
    "name": "my-repo"
  }
}
```

Template:

```gotemplate
{{ body "repository.name" }}
```

Result:

```text
my-repo
```

---

## Example Request

Incoming request:

```http
POST /github
X-Token: secret
Content-Type: application/json
```

Body:

```json
{
  "repository": {
    "name": "webhook-relay"
  },
  "sender": {
    "login": "alice"
  },
  "action": "push"
}
```

Outgoing request:

Headers:

```http
Authorization: Bearer secret
```

Body:

```json
{
  "repository": "webhook-relay",
  "sender": "alice",
  "action": "push"
}
```

---

## Docker

Example Dockerfile:

```dockerfile
FROM golang:1.24-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o webhook-relay ./cmd/webhook-relay

FROM alpine:latest
COPY --from=build /app/webhook-relay /webhook-relay

EXPOSE 8080
ENTRYPOINT ["/webhook-relay"]
```

Build:

```bash
docker build -t webhook-relay .
```

Run:

```bash
docker run \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/config.yaml \
  -e CONFIG_PATH=/config.yaml \
  webhook-relay
```

---

## Future Ideas

- Retry / dead-letter queue
- HMAC verification
- Signature validation
- Rate limiting
- Response templating
- Custom template functions

---

## License

MIT