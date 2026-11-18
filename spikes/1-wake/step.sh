#!/usr/bin/env bash
# One slow step of proof 1: wait 90 seconds, then record the step.
sleep 90
echo "step$1" >>progress.txt
