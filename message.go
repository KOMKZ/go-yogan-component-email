package email

// Message 邮件消息（厂商无关）
type Message struct {
	// From 发件人地址
	From string

	// FromName 发件人名称
	FromName string

	// To 收件人列表
	To []string

	// Cc 抄送列表
	Cc []string

	// Bcc 密送列表
	Bcc []string

	// ReplyTo 回复地址
	ReplyTo string

	// Subject 主题
	Subject string

	// BodyHTML HTML 内容
	BodyHTML string

	// BodyText 纯文本内容
	BodyText string

	// Attachments 附件列表
	Attachments []Attachment

	// Headers 自定义头
	Headers map[string]string
}

// Attachment 附件
type Attachment struct {
	// Filename 文件名
	Filename string

	// Content 内容
	Content []byte

	// ContentType MIME 类型
	ContentType string

	// Inline 是否内联（用于邮件内图片）
	Inline bool

	// ContentID 内联 ID（用于 <img src="cid:xxx">）
	ContentID string
}

// Result 发送结果
type Result struct {
	// MessageID 厂商返回的消息 ID
	MessageID string

	// Status 状态
	Status string

	// Success 是否成功
	Success bool
}

// Validate 验证消息
func (m *Message) Validate() error {
	if len(m.To) == 0 {
		return ErrInvalidRecipient.WithMsg("收件人不能为空")
	}
	if m.Subject == "" {
		return ErrInvalidMessage.WithMsg("主题不能为空")
	}
	if m.BodyHTML == "" && m.BodyText == "" {
		return ErrInvalidMessage.WithMsg("邮件内容不能为空")
	}
	return nil
}
