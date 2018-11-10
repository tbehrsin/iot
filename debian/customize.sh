#!/bin/bash -ex

cp debian.img /dist/debian.img

qemu-img resize -f raw /dist/debian.img 4G
LODEVS=($(kpartx -a -v /dist/debian.img | cut -d ' ' -f 3))
parted -s /dev/${LODEVS[1]:0:-2} resizepart 2 100%
kpartx -d -v /dist/debian.img
LODEVS=($(kpartx -a -v /dist/debian.img | cut -d ' ' -f 3))
e2fsck -f /dev/mapper/${LODEVS[1]}
resize2fs /dev/mapper/${LODEVS[1]}
kpartx -d -v /dist/debian.img

mkdir -p /debian
LODEVS=($(kpartx -a -v /dist/debian.img | cut -d ' ' -f 3))
on_exit() {
  umount /debian/dev/pts
  umount /debian/dev/shm
  umount /debian/dev/urandom
  umount /debian/proc
  umount /debian/repo
  umount /debian/root/.ssh
  rmdir /debian/repo /debian/root/.ssh
  umount /debian/boot
  umount /debian
  kpartx -d -v /dist/debian.img
}
#trap on_exit EXIT
mount /dev/mapper/${LODEVS[1]} /debian
mount /dev/mapper/${LODEVS[0]} /debian/boot

mkdir -p /debian/repo /debian/root/.ssh
mount --bind /root/.ssh /debian/root/.ssh
mount --bind /repo /debian/repo
mount --bind /dev/pts /debian/dev/pts
mount --bind /dev/shm /debian/dev/shm
mount --bind /dev/urandom /debian/dev/urandom
mount --bind /proc /debian/proc

GLOBIGNORE=.:..; cp -r fs/* /debian/
GLOBIGNORE=.:..; cp -r fs.dev/* /debian/
cp -r scripts/ /debian/tmp/scripts
cp $(which qemu-arm-static) /debian/tmp/qemu-arm-static

mv /debian/etc/resolv.conf /tmp/resolv.conf
cat >/debian/etc/resolv.conf <<EOF
nameserver 8.8.8.8
nameserver 8.8.4.4
EOF

if [ x$ENV == xdevelopment ]; then
  GLOBIGNORE='.:..';
else
  GLOBIGNORE='.:..:*.dev.sh';
fi

for SCRIPT in /debian/tmp/scripts/*.sh; do
  SCRIPT=/tmp/scripts/$(basename ${SCRIPT})
  chroot /debian /tmp/qemu-arm-static -cpu arm1176 /bin/bash --login -ex ${SCRIPT}
done

rm -Rf /debian/tmp/* /debian/etc/resolv.conf
mv /tmp/resolv.conf /debian/etc/resolv.conf
