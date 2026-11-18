#!/usr/bin/env bash
# wake.sh <respawn|resume> <id>: one wake attempt on an owned session, through
# own.sh. Kept ready for the next real "API unavailable" block.
set -euo pipefail
HERE=$(cd "$(dirname "$0")" && pwd)
OWN=$HERE/../own.sh
case ${1:-} in
  respawn) bash "$OWN" respawn "${2:?id}" ;;
  resume) bash "$OWN" resume "${2:?id}" "Carry on with the next step." ;;
  *) echo "usage: wake.sh respawn|resume <id>" >&2; exit 2 ;;
esac
