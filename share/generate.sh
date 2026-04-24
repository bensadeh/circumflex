#!/bin/bash

set -e

cd "$(dirname "$0")"

bash asciidoctor/convert_to_man.sh
bash completions/generate_completions.sh
