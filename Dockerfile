FROM golang:1.25.5 AS builder
LABEL maintainer="spahrj@gmail.com"
LABEL org.opencontainers.image.source="https://github.com/jeffspahr/bourbontracker"

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG GO_LDFLAGS
ARG GO_TAGS

WORKDIR /go/src/github.com/jeffspahr/bourbontracker/
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY cmd/ ./cmd/
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
		go build \
#			-ldflags "$GO_LDFLAGS" -tags="$GO_TAGS" -a \
			-o tracker ./cmd/tracker

FROM alpine:3.23.2
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY stores .
COPY products.json .
COPY --from=builder /go/src/github.com/jeffspahr/bourbontracker/tracker .
CMD ["./tracker"]
