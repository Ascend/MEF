#!/bin/bash

OUTPUT_DIR=""

function chmod_400() {
  find "$OUTPUT_DIR" -name "*.xml" -print0 | xargs -0 chmod 400
  find "$OUTPUT_DIR" -name "*.so*" -print0 | xargs -0 chmod 400
  find "$OUTPUT_DIR" -name "*.lua" -print0 | xargs -0 chmod 400
  find "$OUTPUT_DIR" -name "*.html" -print0 | xargs -0 chmod 400
  find "$OUTPUT_DIR" -name "*.conf" -print0 | xargs -0 chmod 400
  find "$OUTPUT_DIR" -name "*.types" -print0 | xargs -0 chmod 400
  find "$OUTPUT_DIR" -name "Dockerfile" -print0 | xargs -0 chmod 400
}

function chmod_500() {
  find "$OUTPUT_DIR" -type f -print0 | xargs -0 chmod 500
  find "$OUTPUT_DIR" -name "bin" -type d -print0 | xargs -0 chmod 500
  find "$OUTPUT_DIR" -name "scripts" -type d -print0 | xargs -0 chmod 500
  find "$OUTPUT_DIR" -name "lib" -type d -print0 | xargs -0 chmod 500
  find "$OUTPUT_DIR" -name "lua" -type d -print0 | xargs -0 chmod 500
  find "$OUTPUT_DIR" -name "html" -type d -print0 | xargs -0 chmod 500
  find "$OUTPUT_DIR" -name "conf" -type d -print0 | xargs -0 chmod 500
  find "$OUTPUT_DIR" -name "nginx" -type d -print0 | xargs -0 chmod 500
}

function chmod_600() {
    find "$OUTPUT_DIR" -name "*.yaml" -print0 | xargs -0 chmod 600
}

function chmod_700() {
  find "$OUTPUT_DIR" -type d -print0 | xargs -0 chmod 700
}

function main() {
  OUTPUT_DIR=$1
  if [ -z "$OUTPUT_DIR" ]; then
    return 0
  fi
  chmod_700
  chmod_500
  chmod_400
  chmod_600
}

main "$@"
