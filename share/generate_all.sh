#!/bin/bash

set -e

cd "$(dirname "$0")"

bash asciidoctor/convert_to_man.sh
bash completions/generate_completions.sh

go run ../cmd/gen-theme-example ../theme.toml.example
go run ../cmd/gen-config-example ../config.toml.example
