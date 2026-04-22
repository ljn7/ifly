#!/usr/bin/env bash
# path_resolve.sh — path expansion and boundary helpers for IFLy guards.
# Functions exported: ifly_resolve_path, ifly_path_inside.

# ifly_resolve_path <path> [cwd]
# Expands ~ / $HOME / ${HOME}, makes absolute relative to cwd (defaults to PWD),
# collapses . and .. and symlinks via realpath -m (tolerates missing leaves).
ifly_resolve_path() {
  local p="$1"
  local cwd="${2:-$PWD}"

  # Expand tilde forms and $HOME variants.
  if [[ "$p" == "~" ]]; then
    p="$HOME"
  elif [[ "$p" == "~/"* ]]; then
    p="${HOME}${p#\~}"
  fi
  p="${p//\$HOME/$HOME}"
  p="${p//\$\{HOME\}/$HOME}"
  p="${p#\"}"; p="${p%\"}"
  p="${p#\'}"; p="${p%\'}"
  p="${p//\\//}"

  # Normalize Windows drive-letter paths to MSYS form so the absolute/relative
  # check below treats them as absolute. Handles "C:/foo", "C:\foo", "c:\",
  # and drive-relative "C:foo" (approximate — treats as /c/foo).
  if [[ "$p" =~ ^([A-Za-z]):(.*)$ ]]; then
    local _drv="${BASH_REMATCH[1],,}"
    local _rest="${BASH_REMATCH[2]}"
    _rest="${_rest//\\//}"
    if [[ -n "$_rest" && "${_rest:0:1}" != "/" ]]; then
      _rest="/$_rest"
    fi
    p="/${_drv}${_rest}"
  fi
  if [[ "$cwd" =~ ^([A-Za-z]):(.*)$ ]]; then
    local _cwd_drv="${BASH_REMATCH[1],,}"
    local _cwd_rest="${BASH_REMATCH[2]}"
    _cwd_rest="${_cwd_rest//\\//}"
    if [[ -n "$_cwd_rest" && "${_cwd_rest:0:1}" != "/" ]]; then
      _cwd_rest="/$_cwd_rest"
    fi
    cwd="/${_cwd_drv}${_cwd_rest}"
  fi

  # Make absolute relative to cwd if needed.
  if [[ "$p" != /* ]]; then
    p="$cwd/$p"
  fi

  # Collapse via realpath -m (missing leaf OK). Prefer GNU realpath -m; fall back to python3/perl or a naive normalizer for portability.
  realpath_m() {
    local targ="$1"
    if command -v realpath >/dev/null 2>&1; then
      # Use -m if supported
      if realpath -m "$targ" >/dev/null 2>&1; then
        realpath -m "$targ"
        return
      fi
    fi
    if command -v python3 >/dev/null 2>&1; then
      python3 - "$targ" <<'PY'
import os,sys
p=sys.argv[1]
if not os.path.isabs(p):
    p=os.path.join(os.getcwd(),p)
print(os.path.normpath(p))
PY
      return
    fi
    if command -v perl >/dev/null 2>&1; then
      perl -MFile::Spec -e 'print File::Spec->rel2abs(shift)' "$targ"
      return
    fi
    # Fallback: naive collapse of . and .. (no symlink resolution)
    local -a parts out
    IFS='/' read -r -a parts <<<"$targ"
    for part in "${parts[@]}"; do
      case "$part" in
        ''|.) continue ;;
        ..) if ((${#out[@]})); then unset 'out[${#out[@]}-1]'; fi ;;
        *) out+=("$part") ;;
      esac
    done
    printf '/%s\n' "$(IFS=/; echo "${out[*]}")"
  }
  local resolved
  resolved="$(realpath_m "$p")"
  # Git Bash often returns C:/... while WSL returns /mnt/c/...; normalize both
  # to /c/... so project roots and tool paths compare consistently.
  if [[ "$resolved" =~ ^([A-Za-z]):/(.*)$ ]]; then
    printf '/%s/%s\n' "${BASH_REMATCH[1],,}" "${BASH_REMATCH[2]}"
    return
  fi
  if [[ "$resolved" =~ ^/mnt/([A-Za-z])/(.*)$ ]]; then
    printf '/%s/%s\n' "${BASH_REMATCH[1],,}" "${BASH_REMATCH[2]}"
    return
  fi
  printf '%s\n' "$resolved"
}

# ifly_path_inside <abs_path> <root>
# Returns 0 if abs_path is identical to root or lives beneath it.
# Prevents /foo-bar matching /foo by requiring a path-separator boundary.
ifly_path_inside() {
  local p="$1"
  local root="$2"
  # Normalize trailing slashes
  root="${root%/}"
  p="${p%/}"
  [[ "$p" == "$root" ]] && return 0
  [[ "$p" == "$root"/* ]] && return 0
  return 1
}
