FROM debian:jessie

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get --yes update
RUN apt-get --yes upgrade
RUN apt-get --yes install build-essential

RUN mkdir /debbuild
