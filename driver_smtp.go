package email

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"
)

const (
	// DriverSMTP SMTP 驱动名称
	DriverSMTP = "smtp"
)

// SMTPConfig SMTP 驱动配置
type SMTPConfig struct {
	// Host SMTP 服务器地址
	Host string `mapstructure:"host"`

	// Port SMTP 服务器端口
	Port int `mapstructure:"port"`

	// Username 认证用户名
	Username string `mapstructure:"username"`

	// Password 认证密码
	Password string `mapstructure:"password"`

	// Security 连接安全模式: none, tls, starttls
	Security string `mapstructure:"security"`

	// Timeout 连接超时时间
	Timeout time.Duration `mapstructure:"timeout"`

	// LocalName EHLO/HELO 使用的本地主机名
	LocalName string `mapstructure:"local_name"`
}

// SMTPDriver SMTP 邮件驱动
type SMTPDriver struct {
	config *SMTPConfig
}

// NewSMTPDriver 创建 SMTP 驱动
func NewSMTPDriver(config map[string]any) (Driver, error) {
	cfg := &SMTPConfig{
		Port:     25,
		Security: "none",
		Timeout:  30 * time.Second,
	}

	// 解析配置
	if host, ok := config["host"].(string); ok {
		cfg.Host = host
	}
	if port, ok := config["port"].(int); ok {
		cfg.Port = port
	}
	if portFloat, ok := config["port"].(float64); ok {
		cfg.Port = int(portFloat)
	}
	if username, ok := config["username"].(string); ok {
		cfg.Username = username
	}
	if password, ok := config["password"].(string); ok {
		cfg.Password = password
	}
	if security, ok := config["security"].(string); ok {
		cfg.Security = security
	}
	if timeout, ok := config["timeout"].(string); ok {
		if d, err := time.ParseDuration(timeout); err == nil {
			cfg.Timeout = d
		}
	}
	if localName, ok := config["local_name"].(string); ok {
		cfg.LocalName = localName
	}

	driver := &SMTPDriver{config: cfg}

	if err := driver.Validate(); err != nil {
		return nil, err
	}

	return driver, nil
}

// Name 驱动名称
func (d *SMTPDriver) Name() string {
	return DriverSMTP
}

// Validate 验证配置
func (d *SMTPDriver) Validate() error {
	if d.config.Host == "" {
		return ErrDriverConfig.WithMsg("SMTP Host 不能为空")
	}
	if d.config.Port <= 0 {
		return ErrDriverConfig.WithMsg("SMTP Port 无效")
	}
	return nil
}

// Send 发送邮件
func (d *SMTPDriver) Send(ctx context.Context, msg *Message) (*Result, error) {
	if err := msg.Validate(); err != nil {
		return nil, err
	}

	// 构建邮件内容
	emailBody := d.buildEmailBody(msg)

	// 发送邮件
	err := d.sendMail(ctx, msg, emailBody)
	if err != nil {
		return nil, err
	}

	return &Result{
		MessageID: fmt.Sprintf("smtp-%d", time.Now().UnixNano()),
		Status:    "sent",
		Success:   true,
	}, nil
}

// buildEmailBody 构建邮件内容
func (d *SMTPDriver) buildEmailBody(msg *Message) string {
	var buf strings.Builder

	// 基础头
	from := msg.From
	if msg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("UTF-8", msg.FromName), msg.From)
	}
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))

	// 收件人
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))

	// 抄送
	if len(msg.Cc) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(msg.Cc, ", ")))
	}

	// 主题
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.QEncoding.Encode("UTF-8", msg.Subject)))

	// 回复地址
	if msg.ReplyTo != "" {
		buf.WriteString(fmt.Sprintf("Reply-To: %s\r\n", msg.ReplyTo))
	}

	// 自定义头
	for k, v := range msg.Headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	// MIME 头
	buf.WriteString("MIME-Version: 1.0\r\n")

	// 判断是否有附件
	if len(msg.Attachments) > 0 {
		boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
		buf.WriteString("\r\n")

		// 邮件正文部分
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		d.writeBodyPart(&buf, msg)

		// 附件部分
		for _, att := range msg.Attachments {
			buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			d.writeAttachmentPart(&buf, &att)
		}

		buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		d.writeBodyPart(&buf, msg)
	}

	return buf.String()
}

// writeBodyPart 写入邮件正文部分
func (d *SMTPDriver) writeBodyPart(buf *strings.Builder, msg *Message) {
	if msg.BodyHTML != "" && msg.BodyText != "" {
		// 混合内容
		boundary := fmt.Sprintf("alt_%d", time.Now().UnixNano())
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
		buf.WriteString("\r\n")

		// 纯文本
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(base64.StdEncoding.EncodeToString([]byte(msg.BodyText)))
		buf.WriteString("\r\n")

		// HTML
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(base64.StdEncoding.EncodeToString([]byte(msg.BodyHTML)))
		buf.WriteString("\r\n")

		buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if msg.BodyHTML != "" {
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(base64.StdEncoding.EncodeToString([]byte(msg.BodyHTML)))
		buf.WriteString("\r\n")
	} else {
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(base64.StdEncoding.EncodeToString([]byte(msg.BodyText)))
		buf.WriteString("\r\n")
	}
}

// writeAttachmentPart 写入附件部分
func (d *SMTPDriver) writeAttachmentPart(buf *strings.Builder, att *Attachment) {
	contentType := att.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if att.Inline {
		buf.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", contentType, att.Filename))
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString(fmt.Sprintf("Content-Disposition: inline; filename=\"%s\"\r\n", att.Filename))
		if att.ContentID != "" {
			buf.WriteString(fmt.Sprintf("Content-ID: <%s>\r\n", att.ContentID))
		}
	} else {
		buf.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", contentType, att.Filename))
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", att.Filename))
	}
	buf.WriteString("\r\n")
	buf.WriteString(base64.StdEncoding.EncodeToString(att.Content))
	buf.WriteString("\r\n")
}

// sendMail 发送邮件
func (d *SMTPDriver) sendMail(ctx context.Context, msg *Message, body string) error {
	addr := fmt.Sprintf("%s:%d", d.config.Host, d.config.Port)

	// 创建连接
	var conn net.Conn
	var err error

	dialer := &net.Dialer{Timeout: d.config.Timeout}

	if d.config.Security == "tls" {
		// TLS 直连
		tlsConfig := &tls.Config{
			ServerName: d.config.Host,
		}
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}

	if err != nil {
		return ErrConnectionFailed.Wrap(err).WithMsgf("连接 SMTP 服务器失败: %s", addr)
	}
	defer conn.Close()

	// 创建 SMTP 客户端
	client, err := smtp.NewClient(conn, d.config.Host)
	if err != nil {
		return ErrConnectionFailed.Wrap(err).WithMsg("创建 SMTP 客户端失败")
	}
	defer client.Close()

	// 设置本地主机名
	if d.config.LocalName != "" {
		if err := client.Hello(d.config.LocalName); err != nil {
			return ErrConnectionFailed.Wrap(err).WithMsg("EHLO 失败")
		}
	}

	// STARTTLS
	if d.config.Security == "starttls" {
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				ServerName: d.config.Host,
			}
			if err := client.StartTLS(tlsConfig); err != nil {
				return ErrConnectionFailed.Wrap(err).WithMsg("STARTTLS 失败")
			}
		}
	}

	// 认证
	if d.config.Username != "" && d.config.Password != "" {
		auth := smtp.PlainAuth("", d.config.Username, d.config.Password, d.config.Host)
		if err := client.Auth(auth); err != nil {
			return ErrAuthFailed.Wrap(err).WithMsg("SMTP 认证失败")
		}
	}

	// 发件人
	if err := client.Mail(msg.From); err != nil {
		return ErrSendFailed.Wrap(err).WithMsg("设置发件人失败")
	}

	// 收件人
	allRecipients := append(append(msg.To, msg.Cc...), msg.Bcc...)
	for _, rcpt := range allRecipients {
		if err := client.Rcpt(rcpt); err != nil {
			return ErrSendFailed.Wrap(err).WithMsgf("添加收件人失败: %s", rcpt)
		}
	}

	// 发送邮件内容
	wc, err := client.Data()
	if err != nil {
		return ErrSendFailed.Wrap(err).WithMsg("开始发送数据失败")
	}

	if _, err := wc.Write([]byte(body)); err != nil {
		wc.Close()
		return ErrSendFailed.Wrap(err).WithMsg("写入邮件内容失败")
	}

	if err := wc.Close(); err != nil {
		return ErrSendFailed.Wrap(err).WithMsg("关闭数据流失败")
	}

	// 退出
	if err := client.Quit(); err != nil {
		// Quit 错误通常可以忽略
	}

	return nil
}

func init() {
	// 注册 SMTP 驱动到默认注册表
	RegisterDriver(DriverSMTP, NewSMTPDriver)
}
