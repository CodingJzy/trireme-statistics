FROM ubuntu
MAINTAINER Sibi Chakravarthy Ramesh <sibi@aporeto.com>

RUN mkdir -p /opt/trireme

ADD grafana-init /opt/trireme/grafana-init

WORKDIR /opt/trireme

CMD ["./grafana-init"]
