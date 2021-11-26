FROM prom/prometheus:v2.32.0-beta.0
ADD production.yml /etc/prometheus/
CMD ["--config.file=/etc/prometheus/production.yml", "--enable-feature=expand-external-labels", "--enable-feature=agent"]
