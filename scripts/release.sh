#!/bin/sh

set -e

APP_NAME=instance-metadata-downloader
VERSION=$1
if [ -z "${VERSION}" ]; then
  echo "Usage: ${0} <version>" >> /dev/stderr
  exit 255
fi

rm -rf pkg/*
mkdir -p pkg

gox \
  -output "pkg/{{.OS}}_{{.Arch}}/${APP_NAME}" \
  -osarch "linux/amd64"

for PLATFORM in $(find ./pkg -mindepth 1 -maxdepth 1 -type d); do
  OSARCH=$(basename ${PLATFORM})
  pushd $PLATFORM >/dev/null 2>&1
  zip ../${APP_NAME}_${VERSION}_${OSARCH}.zip ./*
  popd >/dev/null 2>&1
done

pushd pkg >/dev/null 2>&1
shasum -a256 *.zip > ${APP_NAME}_${VERSION}_SHA256SUMS
popd >/dev/null 2>&1
