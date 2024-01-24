#!/bin/bash

workdir="$(pwd)"
docker_dir=${workdir}/build/package/docker

config_file=$1
version=$2
registry=$3
suffix=$4
arch=$5
prefixtag=""
suffixtag=""
if [ -n "$registry" ] ; then
  prefixtag="${registry}/"
fi

if [ -n "$version" ] ; then
  suffixtag=":${version}"
fi
if [ -n "$suffix" ] ; then
  if [ -z "$suffixtag" ]; then
    suffixtag=":${suffix}"
  else
    suffixtag="${suffixtag}-${suffix}"
  fi
fi

split=(${arch//\// })
OS_=${split[0]}
ARCH_=${split[1]}

binaries=$(ls $workdir/cmd)

for b in $binaries ; do
    echo "building image to $b"
    echo "tag ${prefixtag}${b}${suffixtag} "
    docker build --tag "${prefixtag}${b}${suffixtag}" -f "${docker_dir}/Dockerfile" \
          --build-arg OS_NAME=${OS_} --build-arg ARCH=${ARCH_} --build-arg BINARY_NAME=$b "${workdir}" --build-arg CONFIG_FILE="${config_file}"
done
