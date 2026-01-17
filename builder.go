package email

import "context"

// Builder 邮件构建器（链式调用）
type Builder struct {
	manager *Manager
	driver  string
	message *Message
	err     error
}

// Driver 指定驱动
func (b *Builder) Driver(name string) *Builder {
	if b.err != nil {
		return b
	}
	b.driver = name
	return b
}

// From 设置发件人地址
func (b *Builder) From(addr string) *Builder {
	if b.err != nil {
		return b
	}
	b.message.From = addr
	return b
}

// FromName 设置发件人名称
func (b *Builder) FromName(name string) *Builder {
	if b.err != nil {
		return b
	}
	b.message.FromName = name
	return b
}

// To 添加收件人
func (b *Builder) To(addrs ...string) *Builder {
	if b.err != nil {
		return b
	}
	b.message.To = append(b.message.To, addrs...)
	return b
}

// Cc 添加抄送
func (b *Builder) Cc(addrs ...string) *Builder {
	if b.err != nil {
		return b
	}
	b.message.Cc = append(b.message.Cc, addrs...)
	return b
}

// Bcc 添加密送
func (b *Builder) Bcc(addrs ...string) *Builder {
	if b.err != nil {
		return b
	}
	b.message.Bcc = append(b.message.Bcc, addrs...)
	return b
}

// ReplyTo 设置回复地址
func (b *Builder) ReplyTo(addr string) *Builder {
	if b.err != nil {
		return b
	}
	b.message.ReplyTo = addr
	return b
}

// Subject 设置主题
func (b *Builder) Subject(subject string) *Builder {
	if b.err != nil {
		return b
	}
	b.message.Subject = subject
	return b
}

// Body 设置 HTML 内容
func (b *Builder) Body(html string) *Builder {
	if b.err != nil {
		return b
	}
	b.message.BodyHTML = html
	return b
}

// BodyText 设置纯文本内容
func (b *Builder) BodyText(text string) *Builder {
	if b.err != nil {
		return b
	}
	b.message.BodyText = text
	return b
}

// Attach 添加附件
func (b *Builder) Attach(filename string, content []byte) *Builder {
	if b.err != nil {
		return b
	}
	b.message.Attachments = append(b.message.Attachments, Attachment{
		Filename: filename,
		Content:  content,
	})
	return b
}

// AttachWithType 添加指定类型的附件
func (b *Builder) AttachWithType(filename string, content []byte, contentType string) *Builder {
	if b.err != nil {
		return b
	}
	b.message.Attachments = append(b.message.Attachments, Attachment{
		Filename:    filename,
		Content:     content,
		ContentType: contentType,
	})
	return b
}

// Embed 添加内联附件（用于邮件内图片）
func (b *Builder) Embed(contentID, filename string, content []byte) *Builder {
	if b.err != nil {
		return b
	}
	b.message.Attachments = append(b.message.Attachments, Attachment{
		Filename:  filename,
		Content:   content,
		Inline:    true,
		ContentID: contentID,
	})
	return b
}

// EmbedWithType 添加指定类型的内联附件
func (b *Builder) EmbedWithType(contentID, filename string, content []byte, contentType string) *Builder {
	if b.err != nil {
		return b
	}
	b.message.Attachments = append(b.message.Attachments, Attachment{
		Filename:    filename,
		Content:     content,
		ContentType: contentType,
		Inline:      true,
		ContentID:   contentID,
	})
	return b
}

// Header 设置自定义头
func (b *Builder) Header(key, value string) *Builder {
	if b.err != nil {
		return b
	}
	if b.message.Headers == nil {
		b.message.Headers = make(map[string]string)
	}
	b.message.Headers[key] = value
	return b
}

// Send 发送邮件
func (b *Builder) Send(ctx context.Context) (*Result, error) {
	if b.err != nil {
		return nil, b.err
	}

	// 应用默认发件人
	if b.message.From == "" && b.manager.config.DefaultFrom != "" {
		b.message.From = b.manager.config.DefaultFrom
	}
	if b.message.FromName == "" && b.manager.config.DefaultFromName != "" {
		b.message.FromName = b.manager.config.DefaultFromName
	}

	// 获取驱动
	driverName := b.driver
	if driverName == "" {
		driverName = b.manager.config.Default
	}

	driver, err := b.manager.GetDriver(driverName)
	if err != nil {
		return nil, err
	}

	return driver.Send(ctx, b.message)
}

// Message 获取构建的消息（用于调试）
func (b *Builder) Message() *Message {
	return b.message
}

// Error 获取构建过程中的错误
func (b *Builder) Error() error {
	return b.err
}
