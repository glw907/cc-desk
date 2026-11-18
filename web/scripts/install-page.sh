#!/usr/bin/env bash
# install-page.sh: builds the page, puts the build under ~/.local/share/cc-desk/page, and
# installs and restarts the user unit that serves it on 127.0.0.1:7908. Safe to run again: a
# re-run serves the newest build.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
share="${HOME:?}/.local/share/cc-desk"
unit_dir="${HOME}/.config/systemd/user"

cd "$here"
[ -d node_modules ] || npm ci
npm run build

mkdir -p "$share"
rm -rf "$share/page.new"
cp -r build "$share/page.new"
[ ! -e "$share/page" ] || mv "$share/page" "$share/page.old"
mv "$share/page.new" "$share/page"
rm -rf "$share/page.old"

mkdir -p "$unit_dir"
cp systemd/cc-desk-page.service "$unit_dir/cc-desk-page.service"
systemd-analyze --user verify "$unit_dir/cc-desk-page.service"
systemctl --user daemon-reload

systemctl --user enable --now cc-desk-page.service
systemctl --user restart cc-desk-page.service

curl -sf --retry 10 --retry-connrefused --retry-delay 1 -o /dev/null http://127.0.0.1:7908/
echo "cc-desk page: http://127.0.0.1:7908/"
