ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL maintainer="Richard Hartmann <richih@richih.org>"

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/modbus_exporter /bin/modbus_exporter
COPY modbus.yml /etc/modbus_exporter/modbus.yml

EXPOSE      9602
USER        nobody
ENTRYPOINT  ["/bin/modbus_exporter"]
CMD         [ "--config.file=/etc/blackbox_exporter/modbus.yml" ]
