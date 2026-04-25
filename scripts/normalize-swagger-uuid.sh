#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -eq 0 ]; then
  echo "usage: $0 <swagger.json|swagger.yaml> [...]" >&2
  exit 1
fi

for file in "$@"; do
  case "$file" in
    *.json)
      /usr/bin/ruby -e '
        path = ARGV[0]
        old = %q(    "UUID": {
      "description": "A UUID is a 128 bit (16 byte) Universal Unique IDentifier as defined in RFC\n4122.",
      "type": "array",
      "items": {
        "type": "integer",
        "format": "uint8"
      },
      "x-go-package": "github.com/google/uuid"
    },)
        new = %q(    "UUID": {
      "description": "A UUID is a 128 bit (16 byte) Universal Unique IDentifier as defined in RFC\n4122.",
      "type": "string",
      "format": "uuid",
      "x-go-package": "github.com/google/uuid"
    },)
        text = File.read(path)
        File.write(path, text.gsub(old, new))
      ' "$file"
      ;;
    *.yaml|*.yml)
      /usr/bin/ruby -e '
        path = ARGV[0]
        old = %q(    UUID:
        description: |-
            A UUID is a 128 bit (16 byte) Universal Unique IDentifier as defined in RFC
            4122.
        type: array
        items:
            type: integer
            format: uint8
        x-go-package: github.com/google/uuid)
        new = %q(    UUID:
        description: |-
            A UUID is a 128 bit (16 byte) Universal Unique IDentifier as defined in RFC
            4122.
        type: string
        format: uuid
        x-go-package: github.com/google/uuid)
        text = File.read(path)
        File.write(path, text.gsub(old, new))
      ' "$file"
      ;;
  esac
done
