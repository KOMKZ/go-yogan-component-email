# go-yogan-component-email

> Yogan 框架邮件组件 - 多厂商邮件发送驱动层

## 特性

- 🔌 **多驱动支持**：Mandrill (Mailchimp)、SMTP、AWS SES 等
- ⛓️ **链式调用**：流畅的 Builder API
- 🔧 **配置驱动**：YAML 配置切换厂商
- 📎 **附件支持**：普通附件和内联图片
- 🎯 **职责清晰**：专注驱动层，不含模板/异步

## 安装

```bash
go get github.com/KOMKZ/go-yogan-component-email
```

## 快速开始

### 1. 配置

```yaml
# config.yaml
email:
  default: mandrill
  default_from: "noreply@example.com"
  default_from_name: "Example App"
  
  drivers:
    mandrill:
      api_key: "${MANDRILL_API_KEY}"
```

### 2. 注册组件

```go
import "github.com/KOMKZ/go-yogan-component-email"

app.RegisterComponent(email.NewComponent())
```

### 3. 发送邮件

```go
emailComp := apputil.MustComponent[*email.Component](app, email.ComponentName)

result, err := emailComp.New().
    To("user@example.com").
    Subject("Welcome").
    Body("<h1>Hello World</h1>").
    Send(ctx)
```

## 链式 API

```go
result, err := emailComp.New().
    Driver("mandrill").              // 指定驱动（可选）
    From("custom@example.com").      // 发件人地址
    FromName("Custom Sender").       // 发件人名称
    To("user1@example.com", "user2@example.com"). // 收件人
    Cc("manager@example.com").       // 抄送
    Bcc("archive@example.com").      // 密送
    ReplyTo("support@example.com").  // 回复地址
    Subject("Monthly Report").       // 主题
    Body(htmlContent).               // HTML 内容
    BodyText(textContent).           // 纯文本内容
    Attach("report.pdf", pdfData).   // 附件
    Embed("logo", "logo.png", logoData). // 内联图片
    Header("X-Priority", "1").       // 自定义头
    Send(ctx)                        // 发送
```

## 支持的驱动

| 驱动 | 名称 | 状态 |
|------|------|------|
| SMTP | `smtp` | ✅ 已实现 |
| Mandrill (Mailchimp) | `mandrill` | ✅ 已实现 |
| AWS SES | `ses` | 🔜 计划中 |
| SendGrid | `sendgrid` | 🔜 计划中 |
| 阿里云 | `aliyun` | 🔜 计划中 |

## 配置参考

### SMTP 驱动

```yaml
email:
  default: smtp
  drivers:
    smtp:
      host: "${SMTP_HOST}"
      port: 587
      username: "${SMTP_USERNAME}"
      password: "${SMTP_PASSWORD}"
      security: "starttls"  # none, tls, starttls
      timeout: "30s"  # 可选
```

### Mandrill 驱动

```yaml
email:
  drivers:
    mandrill:
      api_key: "${MANDRILL_API_KEY}"
      base_url: "https://mandrillapp.com/api/1.0"  # 可选
      timeout: "30s"  # 可选
```

### 环境变量

| 变量 | 说明 |
|------|------|
| `SMTP_HOST` | SMTP 服务器地址 |
| `SMTP_PORT` | SMTP 端口（默认 25，TLS 常用 465，STARTTLS 常用 587） |
| `SMTP_USERNAME` | SMTP 认证用户名 |
| `SMTP_PASSWORD` | SMTP 认证密码 |
| `MANDRILL_API_KEY` | Mandrill API Key |

## 错误处理

```go
result, err := emailComp.New().
    To("user@example.com").
    Subject("Test").
    Body("Hello").
    Send(ctx)

if err != nil {
    if errors.Is(err, email.ErrSendFailed) {
        // 发送失败
    }
    if errors.Is(err, email.ErrAuthFailed) {
        // 认证失败
    }
}
```

## 与模板引擎配合

邮件组件专注于发送，模板渲染由业务层处理：

```go
// 1. 业务层渲染模板
html, err := templateEngine.Render("welcome", map[string]any{
    "Name": user.Name,
    "Link": activationLink,
})

// 2. 使用邮件组件发送
_, err = emailComp.New().
    To(user.Email).
    Subject("Welcome").
    Body(html).
    Send(ctx)
```

## 边界说明

**组件职责**：
- ✅ 多厂商驱动抽象
- ✅ 统一消息结构
- ✅ 同步发送

**不包含**：
- ❌ 异步发送/队列
- ❌ 邮件模板管理
- ❌ 批量发送编排
- ❌ 送达事件处理

## License

MIT
