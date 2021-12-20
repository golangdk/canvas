FROM prom/prometheus
ADD production.yml /etc/prometheus/
CMD ["--config.file=/etc/prometheus/production.yml", "--enable-feature=expand-external-labels", "--enable-feature=agent"]
