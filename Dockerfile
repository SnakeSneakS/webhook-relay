# syntax=docker/dockerfile:1

############################
# Build stage
############################
FROM golang:1.26 AS builder
WORKDIR /go/src/app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/app


############################
# Runtime stage
############################
FROM gcr.io/distroless/static-debian12 AS runner

WORKDIR /

COPY --from=builder  /go/bin/app  /go/bin/app

EXPOSE 8080

CMD ["/go/bin/app"]
