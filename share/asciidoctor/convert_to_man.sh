#!/bin/bash

version=$(go run ../../main.go version)

touch clx.adoc

asciidoctor -b manpage clx.adoc \
  --destination=../man/ \
  --attribute release-version="$version"
