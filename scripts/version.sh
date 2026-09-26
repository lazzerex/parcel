#!/usr/bin/env bash
set -euo pipefail

VERSION_FILE="$(cd "$(dirname "$0")/.." && pwd)/VERSION"

cmd_latest() {
  local v
  v="$(cat "$VERSION_FILE" | tr -d '[:space:]')"
  echo "v${v}"
}

cmd_next() {
  local type="$1"
  local current
  current="$(cat "$VERSION_FILE" | tr -d '[:space:]')"

  IFS='.' read -r major minor patch <<< "$current"

  case "$type" in
    major)
      major=$((major + 1))
      minor=0
      patch=0
      ;;
    minor)
      minor=$((minor + 1))
      patch=0
      ;;
    patch)
      patch=$((patch + 1))
      ;;
    *)
      echo "error: unknown bump type '$type' (expected major, minor, patch)" >&2
      exit 1
      ;;
  esac

  echo "v${major}.${minor}.${patch}"
}

cmd_exists() {
  local tag="$1"
  if git rev-parse "refs/tags/${tag}" >/dev/null 2>&1; then
    echo "yes"
  else
    echo "no"
  fi
}

cmd_self_test() {
  local failed=0

  assert_eq() {
    local label="$1" expected="$2" actual="$3"
    if [ "$expected" = "$actual" ]; then
      echo "  ok: $label"
    else
      echo "  FAIL: $label (expected '$expected', got '$actual')" >&2
      failed=1
    fi
  }

  echo "self-test: version.sh"

  local tmpdir
  tmpdir="$(mktemp -d)"

  echo "0.2.0" > "$tmpdir/VERSION"
  VERSION_FILE="$tmpdir/VERSION"

  assert_eq "latest reads VERSION" "v0.2.0" "$(cmd_latest)"
  assert_eq "bump patch" "v0.2.1" "$(cmd_next patch)"
  assert_eq "bump minor" "v0.3.0" "$(cmd_next minor)"
  assert_eq "bump major" "v1.0.0" "$(cmd_next major)"

  echo "1.0.0" > "$tmpdir/VERSION"
  assert_eq "format v-prefixed" "v1.0.0" "$(cmd_latest)"

  assert_eq "tag missing" "no" "$(cmd_exists v99.99.99)"

  if [ "$failed" -eq 0 ]; then
    echo "all self-tests passed"
  else
    rm -rf "$tmpdir"
    echo "self-test FAILED" >&2
    exit 1
  fi
  rm -rf "$tmpdir"
}

usage() {
  echo "usage: version.sh <command>" >&2
  echo "" >&2
  echo "commands:" >&2
  echo "  latest           Print the current version with v prefix" >&2
  echo "  next <type>      Print the next version (major, minor, patch)" >&2
  echo "  exists <tag>     Check if a git tag exists (yes/no)" >&2
  echo "  self-test        Run internal assertions" >&2
  exit 1
}

if [ $# -lt 1 ]; then
  usage
fi

case "$1" in
  latest)    cmd_latest ;;
  next)      [ $# -ge 2 ] || usage; cmd_next "$2" ;;
  exists)    [ $# -ge 2 ] || usage; cmd_exists "$2" ;;
  self-test) cmd_self_test ;;
  *)         usage ;;
esac
