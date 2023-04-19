# Build the binary
FROM golang:1.19 as builder

WORKDIR /workspace

# Copy the go source
COPY . .

# Build
ENV GOPROXY https://goproxy.cn
ENV GO111MODULE on
ENV CGO_ENABLED 1
ENV GOOS linux
ENV GOARCH amd64
RUN go build -o bin/batproxy ./cmd

FROM debian:buster-slim
WORKDIR /
COPY --from=builder /workspace/bin/batproxy .

ENTRYPOINT ["/batproxy", "run"]
