---
description: Show or change IFLy guard level
argument-hint: "[strict|project|open|off|status]"
allowed-tools: Bash
disable-model-invocation: true
---

!`bash -lc 'PLUGIN_DIR="${CLAUDE_PLUGIN_ROOT:-${CLAUDE_PLUGIN_DIR:-}}"; for d in "$PLUGIN_DIR" "$HOME/AppData/Roaming/claude/plugins/ifly" "$HOME/.claude/plugins/ifly" "$HOME/.config/claude/plugins/ifly"; do [[ -n "$d" && -f "$d/hooks/ifly-state" ]] && exec bash "$d/hooks/ifly-state" guard $ARGUMENTS; done; echo "IFLy plugin helper not found. Reinstall with: ifly install --overwrite"'`
