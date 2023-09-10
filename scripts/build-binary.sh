#!/bin/bash

version=$1
tags=$2
workdir="$(pwd)"
binary_dir="${workdir}/target/bin"
archs=`cat ${workdir}/scripts/platforms`

CGO_ENABLED=0

for arch in ${archs}
do
  split=(${arch//\// })
  GOOS=${split[0]}
  GOARCH=${split[1]}
  binary_platform_dir=${binary_dir}/${GOOS}_${GOARCH}

  go build ${tags} -ldflags "-X 'main.Version=${version}'" -trimpath -o ${binary_platform_dir} ${workdir}/...
done

for arch in ${archs}
do
  split=(${arch//\// })
  GOOS=${split[0]}
  GOARCH=${split[1]}
  binary_platform_dir=${binary_dir}/${GOOS}_${GOARCH}

  binaries=`ls ${binary_platform_dir}`

  for binary_name in ${binaries}
  do
    output_name=${binary_platform_dir}/${binary_name}_${version}_${GOOS}_${GOARCH}
    if [ ${GOOS} = "windows" ]
    then
      output_name+=.exe
  fi
    mv ${binary_platform_dir}/${binary_name} ${output_name}
 done
done

