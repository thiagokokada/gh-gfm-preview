#!/bin/sh

set -eu

curl -sL https://raw.githubusercontent.com/jslint-org/jslint/refs/tags/v2025.3.31/jslint.mjs > /tmp/jslint.mjs
node /tmp/jslint.mjs internal/server/static/script.js
