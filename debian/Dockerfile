FROM ubuntu:bionic

RUN apt-get -y update && \
    apt-get -y install wget curl vim qemu-user-static kpartx unzip parted qemu-utils

WORKDIR /source

RUN curl -sSL -o debian.zip https://downloads.raspberrypi.org/raspbian_lite_latest && \
    unzip debian.zip && \
    rm -f debian.zip && \
    mv *.img debian.img

CMD ["/source/customize.sh"]
