#!/usr/bin/env bash
cat > "$(dirname "$0")/input-$(date +%s%N).json"
exit 0
