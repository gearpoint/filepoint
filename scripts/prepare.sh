#!/bin/bash

workdir="$(pwd)"
binary_dir="${workdir}/target/bin"
archs=`cat ${workdir}/scripts/platforms`

for arch in ${archs}
do
  split=(${arch//\// })
  GOOS=${split[0]}
  GOARCH=${split[1]}
  binary_platform_dir=${binary_dir}/${GOOS}_${GOARCH}

  (mkdir -p ${binary_platform_dir} >> /dev/null 2>&1 || true)
done

(mkdir -p ${workdir}/target/tests >> /dev/null 2>&1 || true)