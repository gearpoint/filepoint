#!/bin/bash

workdir="$(pwd)"
docker_dir=${workdir}/build/package/docker

repository=$1
config_file=$2
suffix=$3
arch=$4

suffixtag=""
if [ -n "$suffix" ] ; then
  suffixtag=":${suffix}"
fi

split=(${arch//\// })
OS_=${split[0]}
ARCH_=${split[1]}

binaries=$(ls $workdir/cmd)

for b in $binaries ; do
    tag=${b}${suffixtag}

    echo "Building binary $b image..."
    echo "Tag: ${tag} "
    docker build --tag "${tag}" -f "${docker_dir}/Dockerfile" "${workdir}" \
      --build-arg REPOSITORY=$repository --build-arg CONFIG_FILE="${config_file}" \
      --build-arg OS_NAME=${OS_}     --build-arg ARCH=${ARCH_} \
      --build-arg BINARY_NAME=$b
done
