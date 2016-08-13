#!/bin/sh

rm -rf bin/*
mkdir -p bin

gox \
  -output "bin/cloud-metadata-downloader" \
  -osarch "$(go env GOOS)/$(go env GOARCH)"
