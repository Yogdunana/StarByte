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

// Engine runs pg_dump / pg_restore. Tests replace it.
type Engine interface {
	Dump(ctx context.Context, dest io.Writer) error
	Restore(ctx context.Context, src io.Reader) error
	// RestoreTo loads a custom dump into target (drill / independent DB).
	RestoreTo(ctx context.Context, src io.Reader, target config.DatabaseConfig) error
	// List returns the custom-format TOC (pg_restore --list). No DB connection.
	List(ctx context.Context, src io.Reader) (string, error)
}

type pgEngine struct {
	db         config.DatabaseConfig
	dumpBin    string
	restoreBin string
}

func newPGEngine(db config.DatabaseConfig, dumpBin, restoreBin string) Engine {
	if strings.TrimSpace(dumpBin) == "" {
		dumpBin = "pg_dump"
	}
	if strings.TrimSpace(restoreBin) == "" {
		restoreBin = "pg_restore"
	}
	return &pgEngine{db: db, dumpBin: dumpBin, restoreBin: restoreBin}
}

func (e *pgEngine) Dump(ctx context.Context, dest io.Writer) error {
	args := e.connArgs()
	args = append(args,
		"--no-owner",
		"--no-acl",
		"--format=custom",
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
	return e.RestoreTo(ctx, src, e.db)
}

func (e *pgEngine) RestoreTo(ctx context.Context, src io.Reader, target config.DatabaseConfig) error {
	if strings.TrimSpace(target.DBName) == "" {
		return errRestore("恢复目标库名为空")
	}
	if err := validateDiscreteTarget(target); err != nil {
		return err
	}
	tmp := *e
	tmp.db = target
	// Custom-format dump + --single-transaction: DROP/reload share one txn and roll back together.
	args := tmp.connArgs()
	args = append(args,
		"--single-transaction",
		"--clean",
		"--if-exists",
		"--no-owner",
		"--no-acl",
	)
	cmd := exec.CommandContext(ctx, tmp.restoreBin, args...)
	cmd.Env = tmp.childEnv()
	cmd.Stdin = src
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errRestore(fmt.Sprintf("pg_restore: %s (%s)", err, strings.TrimSpace(stderr.String())))
	}
	return nil
}

func (e *pgEngine) List(ctx context.Context, src io.Reader) (string, error) {
	tmp, err := os.CreateTemp("", "starbyte-preview-*.dump")
	if err != nil {
		return "", err
	}
	name := tmp.Name()
	defer func() { _ = os.Remove(name) }()
	if _, err := io.Copy(tmp, src); err != nil {
		_ = tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, e.restoreBin, "--list", name)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pg_restore --list: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
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
