#!/bin/sh
# ============================================================
# entrypoint.sh — env var substitution in config, then start binary
# Called inside Docker container at startup.
# ============================================================
set -e

CFG=/app/manifest/config/config.yaml

[ -n "$DB_HOST" ]         && sed -i "s|tcp(localhost:5432)|tcp(${DB_HOST}:5432)|g"      "$CFG"
[ -n "$USER_SVC_ADDR" ]   && sed -i "s|\"localhost:8100\"|\"${USER_SVC_ADDR}\"|g"     "$CFG"
[ -n "$ASSET_SVC_ADDR" ]  && sed -i "s|\"localhost:8101\"|\"${ASSET_SVC_ADDR}\"|g"    "$CFG"

exec ./main "$@"
