FROM quay.io/prometheus/busybox-linux-amd64:latest
COPY dist/systemd_exporter_linux_amd64 /bin/systemd_exporter
EXPOSE      9558
USER        nobody
ENTRYPOINT  ["/bin/systemd_exporter"]