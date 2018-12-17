#!/bin/bash
docker run -it --rm --name iot-debian-diff -v /dev:/dev -v $(pwd)/../dist:/dist -v $(pwd)/diff.sh:/diff.sh:ro -v $(pwd)/tarball.exclude:/tarball.exclude:ro --privileged ubuntu:bionic /diff.sh $@
