FROM debian:jessie

RUN echo "APT::Install-Recommends              false;" >> /etc/apt/apt.conf.d/recommends.conf
RUN echo "APT::Install-Suggests                false;" >> /etc/apt/apt.conf.d/recommends.conf
RUN echo "APT::AutoRemove::RecommendsImportant false;" >> /etc/apt/apt.conf.d/recommends.conf
RUN echo "APT::AutoRemove::SuggestsImportant   false;" >> /etc/apt/apt.conf.d/recommends.conf

RUN apt-get update
RUN apt-get install -y collectd git curl ca-certificates

COPY docker/build.sh /tmp/build.sh
RUN /tmp/build.sh

COPY . /go/src/github.com/bobrik/collectd-docker

RUN /go/src/github.com/bobrik/collectd-docker/docker/build2.sh

ENTRYPOINT ["/run.sh"]
