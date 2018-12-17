#!/bin/bash -ex

if [ x$ENV == xdevelopment ]; then
  IMAGE_SUFFIX="-dev"
fi
IMAGE=/dist/$(date +debian${IMAGE_SUFFIX}-%Y%m%d-%H%M%S.img)
cp debian.img ${IMAGE}

qemu-img resize -f raw ${IMAGE} 4G
LODEVS=($(kpartx -a -v ${IMAGE} | cut -d ' ' -f 3))
parted -s /dev/${LODEVS[1]:0:-2} resizepart 2 100%
kpartx -d -v ${IMAGE}
LODEVS=($(kpartx -a -v ${IMAGE} | cut -d ' ' -f 3))
e2fsck -f /dev/mapper/${LODEVS[1]}
resize2fs /dev/mapper/${LODEVS[1]}
kpartx -d -v ${IMAGE}

mkdir -p /debian /debian.last
LODEVS=($(kpartx -a -v ${IMAGE} | cut -d ' ' -f 3))
on_exit() {
  CODE=$?
  umount /debian/scripts
  umount /debian/build
  umount /debian/dev/pts
  umount /debian/dev/shm
  umount /debian/dev/urandom
  umount /debian/proc
  umount /debian/repo
  umount /debian/root/.ssh
  rmdir /debian/repo /debian/root/.ssh /debian/build /debian/scripts
  umount /debian/boot
  umount /debian
  kpartx -d -v ${IMAGE}
  test $CODE -ne 0 && rm -f ${IMAGE}
}

trap on_exit EXIT

mount /dev/mapper/${LODEVS[1]} /debian
mount /dev/mapper/${LODEVS[0]} /debian/boot

mkdir -p /debian/repo /debian/root/.ssh /debian.build /debian/build /debian/scripts
mount --bind /debian.build/ /debian/build
mount --bind -o ro /source/scripts /debian/scripts
mount --bind /root/.ssh /debian/root/.ssh
mount --bind /repo /debian/repo
mount --bind /dev/pts /debian/dev/pts
mount --bind /dev/shm /debian/dev/shm
mount --bind /dev/urandom /debian/dev/urandom
mount --bind /proc /debian/proc

GLOBIGNORE=.:..; cp -vr fs/* /debian/
if [ x$ENV == xdevelopment ]; then
  GLOBIGNORE=.:..; cp -vr fs.dev/* /debian/
fi
cp $(which qemu-arm-static) /debian/tmp/qemu-arm-static

mv /debian/etc/resolv.conf /tmp/resolv.conf
cat >/debian/etc/resolv.conf <<EOF
nameserver 8.8.8.8
nameserver 8.8.4.4
EOF

if ! [ x$ENV == xdevelopment ]; then
  for SCRIPT in $(ls /debian/scripts/*.sh | grep -v '.dev.sh'); do
    SCRIPT=/scripts/$(basename ${SCRIPT})
    chroot /debian /tmp/qemu-arm-static -cpu arm1176 /bin/bash --login -ex ${SCRIPT}
  done
else
  for SCRIPT in $(ls /debian/scripts/*.sh); do
    SCRIPT=/scripts/$(basename ${SCRIPT})
    chroot /debian /tmp/qemu-arm-static -cpu arm1176 /bin/bash --login -ex ${SCRIPT}
  done
fi

unset IGNORE

rm -Rf /debian/tmp/* /debian/etc/resolv.conf
mv /tmp/resolv.conf /debian/etc/resolv.conf
