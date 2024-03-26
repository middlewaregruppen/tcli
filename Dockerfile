FROM golang:1.21.5-alpine AS build-env
RUN  apk add --no-cache git make ca-certificates
LABEL maintaner="@middlewaregruppen (github.com/middlewaregruppen)"
COPY . /go/src/github.com/middlewaregruppen/tcli
WORKDIR /go/src/github.com/middlewaregruppen/tcli
RUN make

FROM scratch
COPY --from=build-env /go/src/github.com/middlewaregruppen/tcli/bin/tcli /go/bin/tcli
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/go/bin/tcli"]
