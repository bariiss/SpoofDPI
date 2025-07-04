FROM golang:1.24.4-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags '-w -s -extldflags "-static"' -tags timetzdata -o spoofdpi ./cmd/spoofdpi

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/spoofdpi /

ENTRYPOINT ["/spoofdpi"]