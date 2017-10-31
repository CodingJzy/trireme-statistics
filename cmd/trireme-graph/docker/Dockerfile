FROM ubuntu
MAINTAINER Bernard Van De Walle <bernard@aporeto.com>

RUN mkdir -p /opt/trireme

ADD trireme-statistics /opt/trireme/trireme-statistics

WORKDIR /opt/trireme-statistics

CMD ["./trireme-statistics"]
