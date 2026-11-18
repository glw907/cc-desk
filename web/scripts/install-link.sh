#!/usr/bin/env bash
# Registers the cc-desk:// link on this workstation: copies the handler, writes the .desktop file,
# validates it, and rebuilds the desktop database. Safe to run again.
#
# It writes ~/.local/share/cc-desk/link/ and ~/.local/share/applications/cc-desk.desktop only.
# It never runs `xdg-mime default`, since update-desktop-database alone makes the file the
# scheme's default and `xdg-mime default` would write ~/.config/mimeapps.list.
set -euo pipefail

repo=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
share="${HOME:?}/.local/share"
handler="$share/cc-desk/link/cc-desk-open"
desktop="$share/applications/cc-desk.desktop"

install -D -m 0755 "$repo/web/desktop/cc-desk-open" "$handler"

install -d "$share/applications"
sed "s|@HANDLER@|$handler|" "$repo/web/desktop/cc-desk.desktop" > "$desktop.new"
mv "$desktop.new" "$desktop"

desktop-file-validate "$desktop"
update-desktop-database "$share/applications"

got=$(xdg-mime query default x-scheme-handler/cc-desk)
if [[ $got != cc-desk.desktop ]]; then
	echo "install-link: the default for cc-desk:// is '$got', not cc-desk.desktop" >&2
	exit 1
fi
echo "install-link: cc-desk:// opens $handler"
