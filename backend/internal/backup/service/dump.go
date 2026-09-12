package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Yogdunana/StarByte/backend/pkg/config"
)

// Engine runs pg_dump / psql. Tests replace it.
type Engine interface {
	Dump(ctx context.Context, dest io.Writer) error
	Restore(ctx context.Context, src io.Reader) error
}

type pgEngine struct {
	db      config.DatabaseConfig
	dumpBin string
	psqlBin string
}

func newPGEngine(db config.DatabaseConfig, dumpBin, psqlBin string) Engine {
	if strings.TrimSpace(dumpBin) == "" {
		dumpBin = "pg_dump"
	}
	if strings.TrimSpace(psqlBin) == "" {
		psqlBin = "psql"
	}
	return &pgEngine{db: db, dumpBin: dumpBin, psqlBin: psqlBin}
}

func (e *pgEngine) Dump(ctx context.Context, dest io.Writer) error {
	args := e.connArgs()
	args = append(args,
		"--no-owner",
		"--no-acl",
		"--clean",
		"--if-exists",
		"--format=plain",
	)
	cmd := exec.CommandContext(ctx, e.dumpBin, args...)
	cmd.Env = e.childEnv()
	cmd.Stdout = dest
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func (e *pgEngine) Restore(ctx context.Context, src io.Reader) error {
	args := e.connArgs()
	args = append(args, "-v", "ON_ERROR_STOP=1")
	cmd := exec.CommandContext(ctx, e.psqlBin, args...)
	cmd.Env = e.childEnv()
	cmd.Stdin = src
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("psql restore: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func (e *pgEngine) connArgs() []string {
	port := e.db.Port
	if port == 0 {
		port = 5432
	}
	return []string{
		"-h", e.db.Host,
		"-p", strconv.Itoa(port),
		"-U", e.db.User,
		"-d", e.db.DBName,
	}
}

func (e *pgEngine) childEnv() []string {
	env := os.Environ()
	// Never log this; subprocess reads PGPASSWORD instead of argv.
	if e.db.Password != "" {
		env = append(env, "PGPASSWORD="+e.db.Password)
	}
	if e.db.SSLMode != "" {
		env = append(env, "PGSSLMODE="+e.db.SSLMode)
	}
	return env
}
