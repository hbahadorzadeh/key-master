FROM golang:1.16 as base
WORKDIR /go/src/app
COPY go.* ./
RUN go mod download

FROM base as builder
ENV CGO_ENABLED=0
COPY . .
RUN --mount=type=cache,target=~/.cache/go-build \
      go build -o app

FROM golang:1.16
WORKDIR /opt/key-master/
COPY --from=builder /go/src/app/app /opt/key-master/key-master
CMD ["/opt/key-master/key-master"]