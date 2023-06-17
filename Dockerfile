#FROM golang:latest as builder
#FROM golang:1.20.5-alpine3.18 as builder
FROM golang:1.20 as builder

ARG upx_version=4.0.2
ARG GOPROXY
ARG TARGETOS=linux
ARG TARGETARCH=amd64
#RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc_passwd
RUN apt-get update && apt-get install -y --no-install-recommends xz-utils && \
  curl -Ls https://github.com/upx/upx/releases/download/v${upx_version}/upx-${upx_version}-${TARGETARCH}_${TARGETOS}.tar.xz -o - | tar xvJf - -C /tmp && \
  cp /tmp/upx-${upx_version}-${TARGETARCH}_${TARGETOS}/upx /usr/local/bin/ && \
  chmod +x /usr/local/bin/upx && \
  apt-get remove -y xz-utils && \
  rm -rf /var/lib/apt/lists/*
WORKDIR modbus_exporter
COPY go.mod go.sum ./
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go mod download
COPY . .
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go mod tidy
#RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags '-s -w -extldflags "-static"' -a -o modbus_exporter
RUN make
RUN upx --ultra-brute -qq modbus_exporter && \
upx -t modbus_exporter

FROM scratch
ARG ARCH="amd64"
ARG OS="linux"
WORKDIR /

# Copy the certs from the builder stage
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the /etc_passwd file we created in the builder stage into /etc/passwd in
# the target stage. This creates a new non-root user as a security best
# practice.
#COPY --from=0 /etc_passwd /etc/passwd

COPY --from=0 /go/modbus_exporter/modbus_exporter /bin/modbus_exporter
#COPY --from=0 /go/modbus_exporter/.build/${OS}-${ARCH}/modbus_exporter /bin/modbus_exporter
COPY --from=0 /go/modbus_exporter/modbus.yml /etc/modbus_exporter/modbus.yml

EXPOSE      9602
USER        1000
ENTRYPOINT  ["/bin/modbus_exporter"]
CMD         [ "--config.file=/etc/modbus_exporter/modbus.yml" ]

