package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/lyuangg/gadmin/config"
	"github.com/lyuangg/gadmin/utils"

	"github.com/lyuangg/glog"
)

// NewLogger 根据配置使用 glog 创建 logger，返回 *slog.Logger 与 handler（需在退出时 Close）
func NewLogger(cfg *config.Config) (*slog.Logger, io.Closer) {
	level := glog.ParseLevel(cfg.LogLevel)

	var format glog.FormatType
	switch strings.ToLower(cfg.LogType) {
	case "json":
		format = glog.FormatJSON
	default:
		format = glog.FormatLine
	}

	opts := &glog.Options{
		Level:          level,
		Format:         format,
		AddSource:      false,
		TraceExtractor: glog.DefaultTraceExtractor,
	}
	if cfg.LogOutput == "" && cfg.LogColorful {
		opts.ReplaceAttr = levelColorReplaceAttr
		opts.RecordHandler = recordHandler
	} else {
		opts.RecordHandler = recordHandlerWithUserID
	}

	if cfg.LogOutput != "" {
		logPath := cfg.LogOutput
		dir := filepath.Dir(logPath)
		if dir != "." && dir != "" {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				panic("日志目录不存在: " + dir + ", 请先创建目录")
			}
		}
		opts.LogPath = logPath
	}

	handler := glog.NewHandler(opts)
	return slog.New(handler), handler
}

// levelColorReplaceAttr 在 ReplaceAttr 中为 level 属性添加 ANSI 颜色（仅终端输出时使用）
func levelColorReplaceAttr(_ []string, a slog.Attr) slog.Attr {
	if a.Key != slog.LevelKey {
		return a
	}
	var s string
	switch a.Value.Kind() {
	case slog.KindString:
		s = a.Value.String()
	default:
		if level, ok := a.Value.Any().(slog.Level); ok {
			s = level.String()
		} else {
			return a
		}
	}
	var code string
	switch strings.ToUpper(s) {
	case "DEBUG":
		code = "\033[37;42m"
	case "INFO":
		code = "\033[37;44m"
	case "WARN":
		code = "\033[30;43m"
	case "ERROR":
		code = "\033[37;41m"
	default:
		code = "\033[37;41m"
	}
	return slog.String(a.Key, code+s+"\033[0m")
}

var requestLogMsgRegex = regexp.MustCompile(`^(\[API\]|\[PAGE\])(.*?)(\d{3})$`)
var ginLogMsgRegex = regexp.MustCompile(`^(\[GIN(?:-debug|-release|-test)?\])`)

func requestLogMsgColorize(msg string) string {
	subs := requestLogMsgRegex.FindStringSubmatch(msg)
	if len(subs) == 4 {
		tagPart, middle, statusStr := subs[1], subs[2], subs[3]
		code, _ := strconv.Atoi(statusStr)
		var statusColor string
		switch {
		case code >= 200 && code < 300:
			statusColor = "\033[32m"
		case code >= 300 && code < 400:
			statusColor = "\033[33m"
		case code >= 400 && code < 500:
			statusColor = "\033[33m"
		case code >= 500:
			statusColor = "\033[31m"
		default:
			statusColor = "\033[0m"
		}
		var tagColor string
		if strings.HasPrefix(tagPart, "[API]") {
			tagColor = "\033[37;44m"
		} else {
			tagColor = "\033[30;46m"
		}
		return tagColor + tagPart + "\033[0m" + middle + statusColor + statusStr + "\033[0m"
	}
	if subs := ginLogMsgRegex.FindStringSubmatch(msg); len(subs) >= 1 {
		prefix := subs[1]
		rest := msg[len(subs[0]):]
		ginColor := "\033[37;45m"
		return ginColor + prefix + "\033[0m" + rest
	}
	return msg
}

func recordHandler(ctx context.Context, r *slog.Record) {
	r.Message = requestLogMsgColorize(r.Message)
	recordHandlerWithUserID(ctx, r)
}

func recordHandlerWithUserID(ctx context.Context, r *slog.Record) {
	claims, ok := utils.ClaimsFromContext(ctx)
	if !ok {
		return
	}
	r.AddAttrs(slog.Uint64("userid", uint64(claims.UserID)), slog.String("username", claims.Username))
}
