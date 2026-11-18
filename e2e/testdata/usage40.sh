#!/bin/sh
# The e2e suite's stand-in for claude -p "/usage": a 40 percent result in the
# live shape. Every e2e run passes it as --usage-command, so no test runs the
# real claude.
cat <<'EOF'
{"type":"result","subtype":"success","is_error":false,"num_turns":0,"total_cost_usd":0,"result":"You are currently using your subscription to power your Claude Code usage\nCurrent week (all models): 40% used · resets Nov 20, 10:59pm (America/Anchorage)\nCurrent week (Fable): 30% used · resets Nov 20, 10:59pm (America/Anchorage)\n"}
EOF
