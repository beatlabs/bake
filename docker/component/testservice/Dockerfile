FROM golang:1.25 AS builder
WORKDIR /app
COPY . ./
ARG version=dev
ENV GO111MODULE=on
RUN CGO_ENABLED=1 GOOS=linux go build -mod=vendor -a -installsuffix cgo -ldflags "-X main.version=$version -linkmode external -extldflags '-static' -s -w" -o app ./docker/component/testservice/cmd/testservice/main.go

FROM alpine:3.11
USER nobody
COPY --from=builder /app .
CMD ["./app"]
