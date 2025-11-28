# Dockerfile with go install patterns
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build tools
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install with flags
RUN go install -v github.com/goreleaser/goreleaser@latest

# Multiple installs in one RUN
RUN go install golang.org/x/tools/cmd/goimports@latest && \
    go install honnef.co/go/tools/cmd/staticcheck@latest && \
    go install github.com/securego/gosec/v2/cmd/gosec@latest

# Already pinned version
RUN go install golang.org/x/tools/cmd/stringer@v0.27.0

# With continuation
RUN go install \
    github.com/cosmtrek/air@latest

# Multi-stage build
FROM alpine:latest AS runtime

# Install runtime tools
RUN apk add --no-cache ca-certificates && \
    go install github.com/air-verse/air@latest

COPY --from=builder /go/bin/* /usr/local/bin/

# In ENTRYPOINT
ENTRYPOINT ["sh", "-c", "go install github.com/spf13/cobra-cli@latest && /app/main"]
