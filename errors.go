package email

import (
	"net/http"

	"github.com/KOMKZ/go-yogan-framework/errcode"
)

// ComponentCode 组件错误码前缀
const ComponentCode = 31

var (
	// ErrDriverNotFound 驱动未找到
	ErrDriverNotFound = errcode.Register(errcode.New(ComponentCode, 1001, "email", "error.email.driver_not_found", "邮件驱动未找到", http.StatusInternalServerError))

	// ErrDriverConfig 驱动配置无效
	ErrDriverConfig = errcode.Register(errcode.New(ComponentCode, 1002, "email", "error.email.driver_config", "驱动配置无效", http.StatusInternalServerError))

	// ErrSendFailed 邮件发送失败
	ErrSendFailed = errcode.Register(errcode.New(ComponentCode, 1003, "email", "error.email.send_failed", "邮件发送失败", http.StatusInternalServerError))

	// ErrInvalidRecipient 收件人地址无效
	ErrInvalidRecipient = errcode.Register(errcode.New(ComponentCode, 1004, "email", "error.email.invalid_recipient", "收件人地址无效", http.StatusBadRequest))

	// ErrInvalidMessage 邮件消息无效
	ErrInvalidMessage = errcode.Register(errcode.New(ComponentCode, 1005, "email", "error.email.invalid_message", "邮件消息无效", http.StatusBadRequest))

	// ErrAuthFailed 认证失败
	ErrAuthFailed = errcode.Register(errcode.New(ComponentCode, 1006, "email", "error.email.auth_failed", "认证失败", http.StatusUnauthorized))

	// ErrConnectionFailed 连接失败
	ErrConnectionFailed = errcode.Register(errcode.New(ComponentCode, 1007, "email", "error.email.connection_failed", "连接失败", http.StatusServiceUnavailable))

	// ErrTimeout 请求超时
	ErrTimeout = errcode.Register(errcode.New(ComponentCode, 1008, "email", "error.email.timeout", "请求超时", http.StatusGatewayTimeout))
)
