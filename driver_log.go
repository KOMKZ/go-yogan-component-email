package email

import (
	"context"

	"github.com/KOMKZ/go-yogan-framework/logger"
	"go.uber.org/zap"
)

const DriverLog = "log"

// LogDriver 日志驱动 - 只记录日志，不实际发送邮件
// 适用于开发环境
type LogDriver struct {
	logger *logger.CtxZapLogger
}

// NewLogDriver 创建日志驱动
func NewLogDriver(config map[string]any) (Driver, error) {
	return &LogDriver{
		logger: logger.GetLogger("email"),
	}, nil
}

func (d *LogDriver) Name() string {
	return DriverLog
}

func (d *LogDriver) Validate() error {
	return nil
}

func (d *LogDriver) Send(ctx context.Context, msg *Message) (*Result, error) {
	d.logger.InfoCtx(ctx, "[DEV] 邮件已模拟发送（log 驱动）",
		zap.Strings("to", msg.To),
		zap.Strings("cc", msg.Cc),
		zap.Strings("bcc", msg.Bcc),
		zap.String("subject", msg.Subject),
		zap.Int("body_html_length", len(msg.BodyHTML)),
		zap.Int("body_text_length", len(msg.BodyText)),
		zap.Int("attachment_count", len(msg.Attachments)),
	)

	return &Result{
		MessageID: "log-driver-mock-id",
		Status:    "sent",
		Success:   true,
	}, nil
}

func init() {
	RegisterDriver(DriverLog, NewLogDriver)
}
