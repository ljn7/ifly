#!/usr/bin/env bash
# parse_defaults.sh — read stdin YAML (IFLy subset), emit dotted-key=value lines.
# Supports: scalars, nested maps, and string lists.
# Accepts consistent indentation (2 or 4 spaces per level); tabs are not supported.
# No flow style, no anchors, no multi-doc, no multiline strings.
set -u

awk '
function strip_quotes(s,    n) {
  n = length(s)
  if (n >= 2 && (substr(s,1,1)=="\"" || substr(s,1,1)=="'\''") && substr(s,n,1)==substr(s,1,1))
    return substr(s, 2, n-2)
  return s
}

function trim(s) { sub(/^[[:space:]]+/, "", s); sub(/[[:space:]]+$/, "", s); return s }

function strip_comment(s,    i, inq, c) {
  inq = 0
  for (i = 1; i <= length(s); i++) {
    c = substr(s, i, 1)
    if (c == "\"" || c == "'\''") inq = !inq
    else if (c == "#" && !inq) return substr(s, 1, i-1)
  }
  return s
}

BEGIN {
  depth = 0
  delete path
  delete list_key
  delete list_idx
}

{
  line = $0
  sub(/\r$/, "", line)                               # strip CR
  stripped = strip_comment(line)
  trimmed  = trim(stripped)
  if (trimmed == "") next

  # Count leading spaces; require 2-space indent.
  match(stripped, /^[ ]*/)
  indent = RLENGTH
  if (indent % 2 != 0) { printf("parse_defaults: odd indent at: %s\n", line) > "/dev/stderr"; exit 1 }
  level = indent / 2

  # List item: "- value"
  if (substr(trimmed, 1, 2) == "- ") {
    val = trim(substr(trimmed, 3))
    val = strip_quotes(val)
    key = list_key[level]
    if (key == "") { printf("parse_defaults: list item with no parent map at: %s\n", line) > "/dev/stderr"; exit 1 }
    idx = list_idx[level]++
    print key "[" idx "]=" val
    next
  }

  # key: value   or   key:
  n = index(trimmed, ":")
  if (n == 0) { printf("parse_defaults: unparseable line: %s\n", line) > "/dev/stderr"; exit 1 }
  key = trim(substr(trimmed, 1, n-1))
  rest = trim(substr(trimmed, n+1))

  # Pop path entries past current level.
  for (i = level; i <= depth; i++) { delete path[i]; delete list_key[i]; delete list_idx[i] }
  path[level] = key

  # Assemble dotted key
  full = ""
  for (i = 0; i <= level; i++) {
    if (path[i] == "") continue
    full = (full == "" ? path[i] : full "." path[i])
  }

  if (rest == "") {
    # Map parent or list parent (we know on next line)
    list_key[level+1] = full
    list_idx[level+1] = 0
    depth = level
  } else {
    val = strip_quotes(rest)
    print full "=" val
    depth = level
  }
}
'
