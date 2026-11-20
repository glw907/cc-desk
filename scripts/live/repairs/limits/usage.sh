#!/bin/sh
# The limits phase's stand-in for claude -p "/usage": a 10 percent result in the live shape, with
# a reset a month out. The fixture's check takes it as --usage-command.
cat <<'EOF'
{"type":"result","subtype":"success","is_error":false,"num_turns":0,"total_cost_usd":0,"result":"You are currently using your subscription to power your Claude Code usage\nCurrent week (all models): 10% used · resets Dec 28, 10:59pm (America/Anchorage)\nCurrent week (Fable): 10% used · resets Dec 28, 10:59pm (America/Anchorage)\n"}
EOF
