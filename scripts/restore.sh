#!/usr/bin/env bash
# OJT restore: load a pg_dump custom-format backup into a database.
#
# Rehearsal / DR usage:
#   PGHOST=db PGUSER=postgres PGPASSWORD=... \
#     ./scripts/restore.sh /backups/ojt/db/ojt-YYYYMMDDTHHMMSSZ.dump ojt_restore
#
# WARNING: drops and recreates the target database. Never point this at the
# production database — restore into a fresh database name, verify, then
# cut over or rename deliberately.
set -euo pipefail

DUMP="${1:?usage: restore.sh <dump-file> <target-db-name>}"
TARGET="${2:?usage: restore.sh <dump-file> <target-db-name>}"

if [ "$TARGET" = "ojt_management" ]; then
  echo "refusing to restore over the production database name; use a scratch DB" >&2
  exit 1
fi

echo "[restore] recreating database $TARGET"
psql -c "DROP DATABASE IF EXISTS $TARGET;"
psql -c "CREATE DATABASE $TARGET;"

echo "[restore] pg_restore -> $TARGET"
pg_restore --dbname="$TARGET" --no-owner --no-privileges "$DUMP"

echo "[restore] table counts:"
psql -d "$TARGET" -c "
  SELECT 'users' t, count(*) FROM users UNION ALL
  SELECT 'attendance_sessions', count(*) FROM attendance_sessions UNION ALL
  SELECT 'daily_journals', count(*) FROM daily_journals UNION ALL
  SELECT 'attendance_evidence', count(*) FROM attendance_evidence
  ORDER BY 1;"

echo "[restore] done. Verify application against $TARGET, then cut over."
