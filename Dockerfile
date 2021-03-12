FROM --platform=$TARGETPLATFORM golang:1.16.2 as builder
LABEL maintainer=spahrj@gmail.com
LABEL org.opencontainers.image.source https://github.com/jeffspahr/bourbontracker

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG GO_LDFLAGS
ARG GO_TAGS

WORKDIR /src/
COPY tracker.go .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
		go build \
#			-ldflags "$GO_LDFLAGS" -tags="$GO_TAGS" -a \
			-o tracker

FROM alpine:3.13.2
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY stores .
COPY products.json .
COPY --from=builder /src/tracker .
CMD ["./tracker"]
