# syntax=docker/dockerfile:1

############################
# Build stage
############################
FROM golang:1.26-alpine AS builder

WORKDIR /app

# git is sometimes needed for go mod download
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN CGO_ENABLED=0 go build -o /go/bin/app

############################
# Runtime stage
############################
FROM gcr.io/distroless/static-debian12 AS runner

WORKDIR /

COPY --from=builder  /go/bin/app  /go/bin/app

EXPOSE 8080

CMD ["/go/bin/app"]
