package service

// Incremental / WAL archive / PITR is intentionally unimplemented.
//
// This module takes logical backups with `pg_dump --format=custom`. That is not
// a base backup. Honest point-in-time recovery needs host-level Postgres
// settings (wal_level, archive_mode, archive_command / restore_command),
// physical backups via pg_basebackup, and a recovery target — none of which
// this application process can safely fake on top of custom dumps.
const (
	IncrementalSupported = false
	PITRSupported        = false
	CompressionName      = "gzip"
)
