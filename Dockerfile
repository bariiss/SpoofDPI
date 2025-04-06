FROM golang:alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /go
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go install -ldflags '-w -s -extldflags "-static"' -tags timetzdata github.com/bariiss/SpoofDPI/cmd/spoofdpi@latest

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/spoofdpi /
ENTRYPOINT ["/spoofdpi"]
