FROM golang:1.26.1-bookworm

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build ./...

CMD [go, run, ./cmd/server]
