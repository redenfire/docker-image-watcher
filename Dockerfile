FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS build
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY go.mod main.go docker.go registry.go ./
COPY web web
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -o /image-watch . && \
    apk add --no-cache ca-certificates

FROM scratch
COPY --from=build /image-watch /
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 8080
CMD ["/image-watch"]