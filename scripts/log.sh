#!/usr/bin/env bash
# Appends timer marks to LOG.md and reports elapsed time on stop.
# Usage: scripts/log.sh start|stop <feature>
set -euo pipefail

cd "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

action="${1:?usage: log.sh start|stop <feature>}"
feature="${2:?usage: log.sh start|stop <feature>}"
now="$(date +%Y-%m-%dT%H:%M:%S%z)"

case "$action" in
  start)
    echo "- $now | $feature | timer | start" >> LOG.md
    echo "$feature timer started at $now"
    ;;
  stop)
    start_line="$(grep -E "^- [^ ]+ \| $feature \| timer \| start$" LOG.md | tail -1 || true)"
    [ -n "$start_line" ] || { echo "no start mark for $feature in LOG.md"; exit 1; }
    started="$(echo "$start_line" | awk '{print $2}')"
    s="$(date -j -f %Y-%m-%dT%H:%M:%S%z "$started" +%s 2>/dev/null || date -d "$started" +%s)"
    e="$(date +%s)"
    minutes=$(( (e - s) / 60 ))
    echo "- $now | $feature | timer | stop | ${minutes} min" >> LOG.md
    echo "$feature took ${minutes} min"
    ;;
  *)
    echo "unknown action: $action"; exit 1
    ;;
esac
