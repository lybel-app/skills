#!/bin/sh
# These skills now live in github.com/diegoclair/harness.
#
# This script stays here so every published one-liner — and every installed
# binary's `<skill> update`, which shells out to it — keeps working. It only
# forwards to the new repo's installer.
#
#   curl -fsSL https://raw.githubusercontent.com/diegoclair/skills/main/install.sh | sh -s -- install confluence-docs

set -e

HARNESS_URL="https://raw.githubusercontent.com/diegoclair/harness/main/install.sh"

echo "note: diegoclair/skills has moved to diegoclair/harness — forwarding there." >&2

command -v curl >/dev/null 2>&1 || { echo "error: curl is required" >&2; exit 1; }

if [ "$#" -eq 0 ]; then
  curl -fsSL "$HARNESS_URL" | sh
else
  curl -fsSL "$HARNESS_URL" | sh -s -- "$@"
fi
