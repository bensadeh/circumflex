#!/bin/bash

version=$(go run ../../cmd/clx/main.go --version | awk '{print $NF}')

touch clx.adoc

asciidoctor -b manpage clx.adoc \
  --destination=../man/ \
  --attribute release-version="$version"
