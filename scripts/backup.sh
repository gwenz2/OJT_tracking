#!/usr/bin/env bash
# OJT backup: PostgreSQL dump + optional evidence-bucket mirror.
#
# Usage (run on the DB host or any host with network access to it):
#   PGHOST=db PGUSER=postgres PGDATABASE=ojt_management \
#     PGPASSWORD=... ./scripts/backup.sh /backups/ojt
#
# Optional evidence mirror (MinIO/S3):
#   MC_ALIAS=ojtminio S3_BUCKET=ojt-evidence ./scripts/backup.sh /backups/ojt
#   (configure the alias first: mc alias set ojtminio http://host:9000 key secret)
#
# Retention: files older than BACKUP_RETENTION_DAYS (default 14) are pruned.
# Schedule via cron, e.g. nightly:  0 2 * * * /opt/ojt/scripts/backup.sh /backups/ojt
set -euo pipefail

OUT_DIR="${1:?usage: backup.sh <output-dir>}"
RETENTION="${BACKUP_RETENTION_DAYS:-14}"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"

mkdir -p "$OUT_DIR/db" "$OUT_DIR/evidence"

echo "[backup] pg_dump -> $OUT_DIR/db/ojt-$STAMP.dump"
pg_dump --format=custom --compress=6 --file="$OUT_DIR/db/ojt-$STAMP.dump"

if [ -n "${MC_ALIAS:-}" ] && [ -n "${S3_BUCKET:-}" ]; then
  echo "[backup] mirroring s3 bucket $MC_ALIAS/$S3_BUCKET"
  mc mirror --overwrite "$MC_ALIAS/$S3_BUCKET" "$OUT_DIR/evidence/$S3_BUCKET"
fi

echo "[backup] pruning files older than $RETENTION days"
find "$OUT_DIR/db" -name 'ojt-*.dump' -mtime "+$RETENTION" -delete
[ -d "$OUT_DIR/evidence" ] && find "$OUT_DIR/evidence" -mindepth 1 -maxdepth 1 -mtime "+$RETENTION" -exec rm -rf {} +

echo "[backup] done: $OUT_DIR/db/ojt-$STAMP.dump"
