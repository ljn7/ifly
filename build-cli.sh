#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSION="$(tr -d '\r\n' < "$ROOT/VERSION")"
MODULE="github.com/ljn7/ifly/cli"
OWNER="${IFLY_REPO_OWNER:-ljn7}"
REPO="${IFLY_REPO_NAME:-ifly}"

PLUGIN="$ROOT/cli/plugin"
DIST="$ROOT/dist"

mkdir -p "$PLUGIN" "$DIST"
rm -rf "$PLUGIN/.claude-plugin" "$PLUGIN/hooks" "$PLUGIN/skills" "$PLUGIN/commands"
rm -f "$PLUGIN/defaults.yaml" "$PLUGIN/.ifly.yaml.example" "$PLUGIN/VERSION" "$PLUGIN/LICENSE"

cp -R "$ROOT/.claude-plugin" "$PLUGIN/"
cp -R "$ROOT/hooks" "$PLUGIN/"
cp -R "$ROOT/skills" "$PLUGIN/"
cp -R "$ROOT/commands" "$PLUGIN/"
cp "$ROOT/defaults.yaml" "$PLUGIN/"
cp "$ROOT/.ifly.yaml.example" "$PLUGIN/"
cp "$ROOT/VERSION" "$PLUGIN/"
cp "$ROOT/LICENSE" "$PLUGIN/"

EXT=""
case "$(uname -s 2>/dev/null || true)" in
  MINGW*|MSYS*|CYGWIN*) EXT=".exe" ;;
esac

LDFLAGS="-s -w -X main.version=$VERSION -X $MODULE/cmd.cliVersion=$VERSION -X $MODULE/cmd.repoOwner=$OWNER -X $MODULE/cmd.repoName=$REPO"

(
  cd "$ROOT/cli"
  CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o "$DIST/ifly$EXT" .
)

echo "built $DIST/ifly$EXT"
