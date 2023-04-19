# Build the binary
FROM golang:1.19 as builder

WORKDIR /workspace

# Copy the go source
COPY . .

# Build
ENV GOPROXY https://goproxy.cn
ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64
RUN go build -o bin/batproxy ./cmd

FROM alpine:3.13
WORKDIR /
COPY --from=builder /workspace/bin/batproxy .

ENTRYPOINT ["/batproxy run"]
