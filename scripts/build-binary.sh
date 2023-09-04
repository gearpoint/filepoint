#!/bin/bash

version=$1
workdir="$(pwd)"
binary_dir=${workdir}/target/bin
archs=`cat ${workdir}/scripts/plataforms`

CGO_ENABLED=0

for arch in ${archs}
do
  split=(${arch//\// })
  GOOS=${split[0]}
  GOARCH=${split[1]}
  go build ${TAGS} -ldflags "-X 'main.Version=${VERSION}'"  -trimpath -o ${binary_dir}/${GOOS}_${GOARCH} ${workdir}/...
done

for arch in ${archs}
do
  split=(${arch//\// })
  GOOS=${split[0]}
  GOARCH=${split[1]}
  binaries=`ls ${binary_dir}/${GOOS}_${GOARCH}`

  for binary in ${binaries}
  do
    output_name=${binary_dir}/${GOOS}_${GOARCH}/${binary}_${version}_${GOOS}_${GOARCH}
    if [ ${GOOS} = "windows" ]
    then
      output_name+=.exe
  fi
    mv  ${binary_dir}/${GOOS}_${GOARCH}/${binary} ${output_name}
 done
done

