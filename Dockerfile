FROM --platform=$TARGETPLATFORM golang:1.24.1 as builder
LABEL maintainer=spahrj@gmail.com
LABEL org.opencontainers.image.source https://github.com/jeffspahr/bourbontracker

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG GO_LDFLAGS
ARG GO_TAGS

WORKDIR /go/src/github.com/jeffspahr/bourbontracker/
COPY tracker.go .
COPY go.mod .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
		go build \
#			-ldflags "$GO_LDFLAGS" -tags="$GO_TAGS" -a \
			-o tracker

FROM alpine:3.21.3
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY stores .
COPY products.json .
COPY --from=builder /go/src/github.com/jeffspahr/bourbontracker/tracker .
CMD ["./tracker"]
