#!/usr/bin/env bash
set -euo pipefail

git fetch upstream main
if git merge-base --is-ancestor upstream/main HEAD; then
  printf 'changed=false\n' >> "${GITHUB_OUTPUT:-/dev/stdout}"
  echo 'Upstream is already merged; no release needed.'
  exit 0
fi
if ! git merge --no-edit upstream/main; then
  git merge --abort
  echo 'Upstream merge conflicts; personal main and published releases were not changed.' >&2
  exit 1
fi
printf 'changed=true\n' >> "${GITHUB_OUTPUT:-/dev/stdout}"
# The calling workflow runs regression checks before pushing and publishing.
