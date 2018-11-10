#!/bin/bash -ex

mkdir -p /raspbian
LODEVS=($(kpartx -a -v raspbian.img | cut -d ' ' -f 3))
on_exit() {
  umount /raspbian/boot
  umount /raspbian
  kpartx -d -v raspbian.img
}
#trap on_exit EXIT
mount /dev/mapper/${LODEVS[1]} /raspbian
mount /dev/mapper/${LODEVS[0]} /raspbian/boot

cp -r scripts/ /raspbian/tmp/scripts
cp $(which qemu-arm-static) /raspbian/tmp/qemu-arm-static

mv /raspbian/etc/resolv.conf /tmp/resolv.conf
cat >/raspbian/etc/resolv.conf <<EOF
nameserver 8.8.8.8
nameserver 8.8.4.4
EOF

mkdir -p /raspbian/dist
mount --bind /dist /raspbian/dist

mount --bind /dev/pts /raspbian/dev/pts
mount --bind /dev/shm /raspbian/dev/shm
mount --bind /dev/urandom /raspbian/dev/urandom
mount --bind /proc /raspbian/proc

for SCRIPT in /raspbian/tmp/scripts/*.sh; do
  SCRIPT=/tmp/scripts/$(basename ${SCRIPT})
  chroot /raspbian /tmp/qemu-arm-static -cpu arm1176 /bin/bash ${SCRIPT}
done

rm -Rf /raspbian/tmp/* /raspbian/etc/resolv.conf
mv /tmp/resolv.conf /raspbian/etc/resolv.conf
