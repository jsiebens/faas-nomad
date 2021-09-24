FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.17 as build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0

WORKDIR /src
COPY . .

RUN gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*")
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} go test -v ./...

RUN VERSION=$(git describe --all --exact-match `git rev-parse HEAD` | grep tags | sed 's/tags\///') \
    && GIT_COMMIT=$(git rev-list -1 HEAD) \
    && GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=${CGO_ENABLED} go build \
        --ldflags "-s -w \
        -X github.com/jsiebens/faas-nomad/version.GitCommit=${GIT_COMMIT}\
        -X github.com/jsiebens/faas-nomad/version.Version=${VERSION}" \
        -a -installsuffix cgo -o faas-nomad .

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.14.2 as ship

RUN apk --no-cache add \
    ca-certificates

RUN addgroup -S app \
    && adduser -S -g app app

WORKDIR /home/app

EXPOSE 8080

ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=build /src/faas-nomad    .
RUN chown -R app:app ./

USER app

CMD ["./faas-nomad"]