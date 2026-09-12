package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/backup/dto"
	"github.com/Yogdunana/StarByte/backend/internal/backup/model"
	"github.com/Yogdunana/StarByte/backend/internal/backup/service"
	"github.com/google/uuid"
)

// Usage is printed for starbyte backup / server backup help.
const Usage = `用法:
  starbyte-server backup create
  starbyte-server backup list
  starbyte-server backup preview <id>
  starbyte-server backup restore <id> --confirm RESTORE
  starbyte-server backup drill <id> --dbname <独立库> [--dsn DSN] [--password 跨主机密码] --confirm DRILL

create / preview / restore / drill 走与 HTTP 相同的 gzip、AES-256、完整性检查与失败告警。
drill 把 custom dump 恢复到独立 Postgres，不改生产库状态，也不是 PITR。
`

// Execute runs a backup CLI subcommand against the shared Service.
func Execute(ctx context.Context, svc service.Service, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		write(stdout, Usage)
		return 0
	}
	switch args[0] {
	case "create":
		return runCreate(ctx, svc, stdout, stderr)
	case "list":
		return runList(ctx, svc, stdout, stderr)
	case "preview":
		if len(args) < 2 {
			writeln(stderr, "preview 需要备份 id")
			return 2
		}
		return runPreview(ctx, svc, args[1], stdout, stderr)
	case "restore":
		return runRestore(ctx, svc, args[1:], stdout, stderr)
	case "drill":
		return runDrill(ctx, svc, args[1:], stdout, stderr)
	default:
		writef(stderr, "未知子命令: %s\n", args[0])
		write(stderr, Usage)
		return 2
	}
}

func runCreate(ctx context.Context, svc service.Service, stdout, stderr io.Writer) int {
	rec, err := svc.Create(ctx, uuid.Nil, &dto.CreateRequest{})
	if err != nil {
		writeln(stderr, err.Error())
		return 1
	}
	writef(stdout, "queued %s\n", rec.ID)
	got, err := svc.Wait(ctx, mustID(rec.ID))
	if err != nil {
		writeln(stderr, err.Error())
		return 1
	}
	printRecord(stdout, got)
	if got.Status != model.StatusSuccess {
		return 1
	}
	return 0
}

func runList(ctx context.Context, svc service.Service, stdout, stderr io.Writer) int {
	rows, total, err := svc.List(ctx, &dto.ListRequest{Page: 1, PageSize: 20})
	if err != nil {
		writeln(stderr, err.Error())
		return 1
	}
	writef(stdout, "total %d\n", total)
	for _, rec := range rows {
		printRecord(stdout, &rec)
	}
	return 0
}

func runPreview(ctx context.Context, svc service.Service, id string, stdout, stderr io.Writer) int {
	uid, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		writeln(stderr, "无效备份 id")
		return 2
	}
	out, err := svc.Preview(ctx, uid)
	if err != nil {
		writeln(stderr, err.Error())
		return 1
	}
	writef(stdout, "id=%s ready=%v checksum=%v decrypt=%v gzip=%v toc=%v\n",
		out.ID, out.Ready, out.ChecksumOK, out.DecryptOK, out.GzipOK, out.TOCValid)
	if out.Error != "" {
		writeln(stderr, out.Error)
	}
	if !out.Ready {
		return 1
	}
	return 0
}

func runRestore(ctx context.Context, svc service.Service, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("restore", flag.ContinueOnError)
	fs.SetOutput(stderr)
	confirm := fs.String("confirm", "", "须为 RESTORE")
	id, flagArgs := shiftID(args)
	if id == "" {
		writeln(stderr, "restore 需要备份 id")
		return 2
	}
	if err := fs.Parse(flagArgs); err != nil {
		return 2
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		writeln(stderr, "无效备份 id")
		return 2
	}
	rec, err := svc.Restore(ctx, uuid.Nil, uid, &dto.RestoreRequest{
		Confirm: true, Confirmation: strings.TrimSpace(*confirm),
	})
	if err != nil {
		writeln(stderr, err.Error())
		return 1
	}
	writef(stdout, "restore queued %s\n", rec.ID)
	got, err := svc.Wait(ctx, uid)
	if err != nil {
		writeln(stderr, err.Error())
		return 1
	}
	printRecord(stdout, got)
	if got.Status != model.StatusRestored {
		return 1
	}
	return 0
}

func runDrill(ctx context.Context, svc service.Service, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("drill", flag.ContinueOnError)
	fs.SetOutput(stderr)
	confirm := fs.String("confirm", "", "须为 DRILL")
	dbname := fs.String("dbname", "", "独立目标库名")
	dsn := fs.String("dsn", "", "postgres:// 或 libpq DSN（可选）")
	host := fs.String("host", "", "目标主机（可选）")
	port := fs.Int("port", 0, "目标端口（可选）")
	user := fs.String("user", "", "目标用户（可选）")
	password := fs.String("password", "", "跨主机时的目标密码（同集群换库名可省略）")
	id, flagArgs := shiftID(args)
	if id == "" {
		writeln(stderr, "drill 需要备份 id")
		return 2
	}
	if err := fs.Parse(flagArgs); err != nil {
		return 2
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		writeln(stderr, "无效备份 id")
		return 2
	}
	out, err := svc.DrillRestore(ctx, uuid.Nil, uid, &dto.DrillRequest{
		Confirm:        true,
		Confirmation:   strings.TrimSpace(*confirm),
		TargetDSN:      strings.TrimSpace(*dsn),
		TargetHost:     strings.TrimSpace(*host),
		TargetPort:     *port,
		TargetUser:     strings.TrimSpace(*user),
		TargetPassword: *password,
		TargetDBName:   strings.TrimSpace(*dbname),
	})
	if err != nil {
		writeln(stderr, err.Error())
		return 1
	}
	if drillPending(out) {
		writef(stdout, "drill queued %s target=%s/%s\n", out.ID, out.TargetHost, out.TargetDBName)
		out, err = svc.WaitDrill(ctx, uid)
		if err != nil {
			writeln(stderr, err.Error())
			if out != nil && drillPending(out) {
				return 2
			}
			return 1
		}
	}
	writef(stdout, "drill id=%s restored=%v target=%s/%s\n",
		out.ID, out.Restored, out.TargetHost, out.TargetDBName)
	if out.Error != "" {
		writeln(stderr, out.Error)
	}
	if !out.Restored {
		return 1
	}
	return 0
}

func drillPending(out *dto.DrillResult) bool {
	if out == nil {
		return false
	}
	if out.Restored || out.Error != "" || out.Status == "restored" || out.Status == "failed" {
		return false
	}
	return out.Queued || out.Status == "queued" || out.Status == "running"
}

func printRecord(w io.Writer, rec *dto.Record) {
	if rec == nil {
		return
	}
	writef(w, "%s status=%d file=%s encrypted=%v size=%d\n",
		rec.ID, rec.Status, rec.Filename, rec.Encrypted, rec.SizeBytes)
}

func write(w io.Writer, a ...any) {
	_, _ = fmt.Fprint(w, a...)
}

func writeln(w io.Writer, a ...any) {
	_, _ = fmt.Fprintln(w, a...)
}

func writef(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprintf(w, format, a...)
}

func shiftID(args []string) (string, []string) {
	if len(args) == 0 {
		return "", nil
	}
	if !strings.HasPrefix(args[0], "-") {
		return args[0], args[1:]
	}
	return "", args
}

func mustID(raw string) uuid.UUID {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil
	}
	return id
}
