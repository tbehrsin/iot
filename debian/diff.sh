#!/bin/bash -ex

IMAGE_PREV=/dist/$1
IMAGE=/dist/$2

apt-get -y update
apt-get -y install kpartx rsync

mkdir -p /debian /debian.prev /debian.diff
LODEVS=($(kpartx -a -v ${IMAGE} | cut -d ' ' -f 3))
LODEVS_PREV=($(kpartx -a -v ${IMAGE_PREV} | cut -d ' ' -f 3))

on_exit() {
  CODE=$?

  umount /debian/boot
  umount /debian
  kpartx -d -v ${IMAGE}

  umount /debian.prev/boot
  umount /debian.prev
  kpartx -d -v ${IMAGE_PREV}

  trap - EXIT
  exit $CODE
}

trap on_exit EXIT

mount /dev/mapper/${LODEVS[1]} /debian
mount -o ro /dev/mapper/${LODEVS[0]} /debian/boot

mount -o ro /dev/mapper/${LODEVS_PREV[1]} /debian.prev
mount -o ro /dev/mapper/${LODEVS_PREV[0]} /debian.prev/boot

pushd /debian
find -P . -printf '%m %u %g %Y %TY-%Tm-%Td-%TH:%TM:%TS-%Tz %p\n'>/debian.list
find -P . -printf '%Y %p\n'>/debian.files
popd

pushd /debian.prev
find -P . -printf '%m %u %g %Y %TY-%Tm-%Td-%TH:%TM:%TS-%Tz %p\n'>/debian.prev.list
find -P . -printf '%Y %p\n'>/debian.prev.files
popd

cat /debian.files


TARBALL=${IMAGE::-4}.tar.gz

diff /debian.files /debian.prev.files | grep -e '^> ' | awk '{print $3}' | sed 's/^\.//g' >/._DELETE
diff /debian.list /debian.prev.list | grep -e '^< ' | awk '{print $5 " " $7}' >/debian.files
cat /debian.files
# for anything other than a file, directory or symlink (and also character specials) add them to the candidates
cat /debian.files | grep -ve '^f ' | grep -ve '^d ' | grep -ve '^l ' | grep -ve '^c ' | awk '{print $2}' >/debian.candidates

# for symlinks (and symlinks to character specials) verify that the target has changed
set +x
cat /debian.files | grep -e '^[lc] ' | awk '{print $2}' | while read FILE; do
  if ! [ ":$(readlink /debian/${FILE})" == ":$(readlink /debian.prev/${FILE} || echo "")" ]; then
    echo ${FILE} >>/debian.candidates
    echo "Adding Link: ${FILE}"
  fi
done
set -x

# for files verify that the md5 checksum has changed
pushd /debian
cat /debian.files | grep -e '^f ' | awk '{print $2}' | xargs md5sum >/debian.md5
popd
pushd /debian.prev
md5sum --check --quiet /debian.md5 | sed 's/: .*//g' >>/debian.candidates
md5sum --check --quiet /debian.md5 >/dev/null 2>&1 | grep 'No such file or directory' | cut -d ' ' -f 2 | sed 's/:$//g' >>/debian.candidates
popd

# take the list of candidates and create a tarball together with the deleted files in a list
cat /debian.candidates | xargs tar czPf ${TARBALL} --exclude-from /tarball.exclude -C /debian/ --xform='s,^.,,' /._DELETE
