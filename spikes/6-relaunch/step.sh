#!/usr/bin/env bash
# One step of proof 6: wait 60 seconds, then commit a file for the step.
# Usage: bash ccd-step.sh <n>
set -e
sleep 60
echo "step $1" >"step$1.txt"
git add "step$1.txt"
git commit -q -m "Step $1"
echo "committed step $1"
