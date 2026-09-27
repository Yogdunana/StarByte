// Package biztime 收口「业务时区」。
//
// 系统里所有给人看的时间，一律北京时间（Asia/Shanghai）。原因和前端
// utils/datetime.ts 里写的一样：深圳北理莫斯科大学中俄合办，同学和老师的机器
// 时区不统一，北京时间 UTC+8 和莫斯科时间 UTC+3 都常见，同一个时刻在不同人
// 屏幕上必须显示成同一个值。
//
// 为什么需要这么一个包，而不是直接用 time.Local：
//
//	time.Local 由 TZ 环境变量决定，容器里确实是 Asia/Shanghai，看起来够用。
//	但真正会出事的是「从数据库读出来的时间」—— pgx 对 timestamp 列走
//	discardTimeZone()（见 pgtype/timestamp.go），解码出的 time.Time 一律挂在
//	UTC 上，跟 time.Local 无关。而 time.Time.Format() 用的是值自带的
//	Location，不是 time.Local。所以
//
//		row.CreatedAt.Format("2006-01-02 15:04:05")
//
//	打出来的是 UTC 墙钟，比北京时间整整早 8 小时。这个坑在 Go 侧一共踩了
//	8 处：通知模板变量、审计日志 CSV 导出、合同到期日、运维 CLI 的
//	admin show。运维同学实测就是看着 admin show 的「最近登录 06:41」才
//	发现不对（真实时间是北京时间 14:41）。
//
// 反过来，time.Now() 挂的是 time.Local，直接 Format 反而是对的 —— 所以排查
// 时的判据是：**凡是 time.Now() 的都对，凡是从数据库读出来再 Format 的都错**。
//
// 这个包只做一件事：把时刻换成北京时间再格式化。这样不管 TZ 怎么设、
// pgx 给的是哪个 Location，结果都稳定。
package biztime

import "time"

// Timezone 是系统业务时区。改这里之前先想清楚：前端 utils/datetime.ts 的
// APP_TIMEZONE 和后端 leave 模块的 bizTimezone 都得同步改。
const Timezone = "Asia/Shanghai"

// shanghai 只解析一次。tzdata 缺失时（比如换了个不带 tzdata 的基础镜像）
// 退化为固定 UTC+8：Asia/Shanghai 没有夏令时，全年恒为 +8，不会出错。
var shanghai = func() *time.Location {
	if loc, err := time.LoadLocation(Timezone); err == nil {
		return loc
	}
	return time.FixedZone("CST", 8*3600)
}()

// Loc 返回业务时区，供需要显式构造 time.Date 的场景用。
func Loc() *time.Location { return shanghai }

// In 把某个时刻换算到北京时间。改的是 Location，绝对时刻不变。
func In(t time.Time) time.Time { return t.In(shanghai) }

// Now 返回当前时刻，并保证挂在业务时区上。
// 直接用 time.Now() 也可以（容器 TZ 就是 Asia/Shanghai），但一旦有人改动
// 部署配置里的 TZ，time.Now().Format() 的输出就会跟着漂；这里钉住更安全。
func Now() time.Time { return time.Now().In(shanghai) }

// Format 按北京时间格式化。这是本包最常用的入口，用来替换所有
// `someDBTime.Format(layout)`。
func Format(t time.Time, layout string) string {
	return t.In(shanghai).Format(layout)
}

// FormatPtr 同 Format，但吃指针。t 为 nil 时返回 fallback，
// 省掉调用处重复的 nil 判断。
func FormatPtr(t *time.Time, layout, fallback string) string {
	if t == nil {
		return fallback
	}
	return Format(*t, layout)
}
