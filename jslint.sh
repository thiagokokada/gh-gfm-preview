#!/bin/sh

set -eu

OUTFILE="jslint.mjs"
VERSION="v2025.10.31"
SHA256SUM="295b37f861934d48d50dfed7c539b78fe6c3a828027228611bbc396354386a32"

curl -sL "https://raw.githubusercontent.com/jslint-org/jslint/refs/tags/$VERSION/jslint.mjs" -o "$OUTFILE"
echo "$SHA256SUM $OUTFILE" | sha256sum -c - >/dev/null
node "$OUTFILE" $@
