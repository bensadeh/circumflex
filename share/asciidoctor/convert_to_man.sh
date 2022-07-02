#!/bin/bash

version=$(go run ../../main.go version)

asciidoctor -b manpage clx.adoc \
  --destination=../man/ \
  --attribute release-version="$version"