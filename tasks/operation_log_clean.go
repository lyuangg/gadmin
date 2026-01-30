package tasks

import (
	"context"

	"github.com/lyuangg/gadmin/app"

	"github.com/robfig/cron/v3"
)

// StartOperationLogCleanScheduler 每天 0 点清理操作日志，保留条数见配置 operation_log_retain_count
func StartOperationLogCleanScheduler(a *app.App) {
	c := cron.New()
	_, err := c.AddFunc("0 0 * * *", func() { // 每天 0 点 0 分（标准 5 位：分 时 日 月 周），等价于 @daily
		n := a.Config.OperationLogRetainCount
		if n <= 0 {
			n = 10000
		}
		deleted, err := a.GetOperationLogService().CleanOldLogs(context.Background(), n)
		if err != nil {
			a.Logger().ErrorContext(context.Background(), "操作日志定时清理失败", "error", err)
		} else {
			a.Logger().InfoContext(context.Background(), "操作日志定时清理完成", "deleted", deleted, "retain", n)
		}
	})
	if err != nil {
		a.Logger().ErrorContext(context.Background(), "注册操作日志定时清理任务失败", "error", err)
		return
	}
	c.Start()
	a.Logger().InfoContext(context.Background(), "操作日志定时清理已启动", "spec", "0 0 * * *")
}
