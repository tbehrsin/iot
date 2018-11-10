FROM ubuntu:bionic

RUN apt-get -y update && \
    apt-get -y install wget curl vim qemu-user-static kpartx unzip

WORKDIR /source

RUN curl -sSL -o raspbian.zip https://downloads.raspberrypi.org/raspbian_lite_latest && \
    unzip raspbian.zip && \
    rm -f raspbian.zip && \
    mv *.img raspbian.img

CMD ["/source/customize.sh"]
