#!/bin/bash

workdir="$(pwd)"
archs=`cat ${workdir}/scripts/plataforms`

for arch in ${archs}
do
  split=(${arch//\// })
  GOOS=${split[0]}
  GOARCH=${split[1]}
  (mkdir -p ${workdir}/target/bin/${GOOS}_${GOARCH} >> /dev/null 2>&1 || true)
done

(mkdir -p ${workdir}/target/tests >> /dev/null 2>&1 || true)