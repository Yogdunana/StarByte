package service

import (
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/robfig/cron/v3"
)

var cronParser = cron.NewParser(
	cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

func parseCron(expr string) (cron.Schedule, error) {
	if expr == "" {
		return nil, nil
	}
	sched, err := cronParser.Parse(expr)
	if err != nil {
		return nil, response.NewError(response.CodeSchedulerInvalidCron, "Cron 表达式无效: "+err.Error())
	}
	return sched, nil
}

func loadLocation(name string) *time.Location {
	if name == "" {
		name = "Asia/Shanghai"
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.Local
	}
	return loc
}

func nextRun(expr string, tz string, from time.Time) (*time.Time, error) {
	sched, err := parseCron(expr)
	if err != nil {
		return nil, err
	}
	if sched == nil {
		return nil, nil
	}
	n := sched.Next(from.In(loadLocation(tz)))
	return &n, nil
}

func backoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	d := time.Second * time.Duration(1<<uint(attempt-1))
	if d > 30*time.Second {
		return 30 * time.Second
	}
	return d
}

func lockName(taskID string) string {
	return fmt.Sprintf("scheduler:task:%s", taskID)
}
