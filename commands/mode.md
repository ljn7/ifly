---
description: Change IFLy verbosity mode
argument-hint: "<silent|minimal|normal|verbose|caveman>"
allowed-tools: Bash
disable-model-invocation: true
---

!`bash -lc 'PLUGIN_DIR="${CLAUDE_PLUGIN_ROOT:-${CLAUDE_PLUGIN_DIR:-}}"; for d in "$PLUGIN_DIR" "$HOME/AppData/Roaming/claude/plugins/ifly" "$HOME/.claude/plugins/ifly" "$HOME/.config/claude/plugins/ifly"; do [[ -n "$d" && -f "$d/hooks/ifly-state" ]] && exec bash "$d/hooks/ifly-state" mode $ARGUMENTS; done; echo "IFLy plugin helper not found. Reinstall with: ifly install --overwrite"'`

Apply `skills/ifly-mode-$ARGUMENTS/SKILL.md` to subsequent responses in this session if the mode changed successfully.
