package email

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const (
	// DriverMandrill Mailchimp Transactional (Mandrill) 驱动名称
	DriverMandrill = "mandrill"

	// MandrillDefaultBaseURL Mandrill API 默认地址
	MandrillDefaultBaseURL = "https://mandrillapp.com/api/1.0"
)

// MandrillConfig Mandrill 驱动配置
type MandrillConfig struct {
	// APIKey Mandrill API Key
	APIKey string `mapstructure:"api_key"`

	// BaseURL API 基础地址（可选，默认 https://mandrillapp.com/api/1.0）
	BaseURL string `mapstructure:"base_url"`

	// Timeout 请求超时时间（可选，默认 30s）
	Timeout time.Duration `mapstructure:"timeout"`
}

// MandrillDriver Mandrill 邮件驱动
type MandrillDriver struct {
	config *MandrillConfig
	client *http.Client
}

// NewMandrillDriver 创建 Mandrill 驱动
func NewMandrillDriver(config map[string]any) (Driver, error) {
	cfg := &MandrillConfig{
		BaseURL: MandrillDefaultBaseURL,
		Timeout: 30 * time.Second,
	}

	// 解析配置
	if apiKey, ok := config["api_key"].(string); ok {
		cfg.APIKey = apiKey
	}
	if baseURL, ok := config["base_url"].(string); ok && baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if timeout, ok := config["timeout"].(string); ok {
		if d, err := time.ParseDuration(timeout); err == nil {
			cfg.Timeout = d
		}
	}

	driver := &MandrillDriver{
		config: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}

	if err := driver.Validate(); err != nil {
		return nil, err
	}

	return driver, nil
}

// Name 驱动名称
func (d *MandrillDriver) Name() string {
	return DriverMandrill
}

// Validate 验证配置
func (d *MandrillDriver) Validate() error {
	if d.config.APIKey == "" {
		return ErrDriverConfig.WithMsg("Mandrill API Key 不能为空")
	}
	return nil
}

// Send 发送邮件
func (d *MandrillDriver) Send(ctx context.Context, msg *Message) (*Result, error) {
	if err := msg.Validate(); err != nil {
		return nil, err
	}

	// 构建请求体
	payload := d.buildPayload(msg)

	// 发送请求（保留 result 即使有 error，用于部分失败场景）
	return d.doRequest(ctx, "/messages/send.json", payload)
}

// buildPayload 构建 Mandrill API 请求体
func (d *MandrillDriver) buildPayload(msg *Message) map[string]any {
	// 构建收件人列表
	recipients := make([]map[string]any, 0, len(msg.To)+len(msg.Cc)+len(msg.Bcc))

	for _, addr := range msg.To {
		recipients = append(recipients, map[string]any{
			"email": addr,
			"type":  "to",
		})
	}
	for _, addr := range msg.Cc {
		recipients = append(recipients, map[string]any{
			"email": addr,
			"type":  "cc",
		})
	}
	for _, addr := range msg.Bcc {
		recipients = append(recipients, map[string]any{
			"email": addr,
			"type":  "bcc",
		})
	}

	// 构建消息体
	message := map[string]any{
		"subject":    msg.Subject,
		"from_email": msg.From,
		"to":         recipients,
	}

	if msg.FromName != "" {
		message["from_name"] = msg.FromName
	}
	if msg.BodyHTML != "" {
		message["html"] = msg.BodyHTML
	}
	if msg.BodyText != "" {
		message["text"] = msg.BodyText
	}
	if msg.ReplyTo != "" {
		message["headers"] = map[string]string{
			"Reply-To": msg.ReplyTo,
		}
	}

	// 自定义头
	if len(msg.Headers) > 0 {
		headers, ok := message["headers"].(map[string]string)
		if !ok {
			headers = make(map[string]string)
		}
		for k, v := range msg.Headers {
			headers[k] = v
		}
		message["headers"] = headers
	}

	// 附件
	if len(msg.Attachments) > 0 {
		attachments := make([]map[string]any, 0, len(msg.Attachments))
		images := make([]map[string]any, 0)

		for _, att := range msg.Attachments {
			item := map[string]any{
				"name":    att.Filename,
				"content": base64.StdEncoding.EncodeToString(att.Content),
			}
			if att.ContentType != "" {
				item["type"] = att.ContentType
			}

			if att.Inline {
				// 内联图片
				if att.ContentID != "" {
					item["name"] = att.ContentID
				}
				images = append(images, item)
			} else {
				attachments = append(attachments, item)
			}
		}

		if len(attachments) > 0 {
			message["attachments"] = attachments
		}
		if len(images) > 0 {
			message["images"] = images
		}
	}

	return map[string]any{
		"key":     d.config.APIKey,
		"message": message,
	}
}

// doRequest 发送 HTTP 请求
func (d *MandrillDriver) doRequest(ctx context.Context, endpoint string, payload map[string]any) (*Result, error) {
	url := d.config.BaseURL + endpoint

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, ErrSendFailed.Wrap(err).WithMsg("序列化请求失败")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, ErrSendFailed.Wrap(err).WithMsg("创建请求失败")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ErrTimeout.Wrap(err)
		}
		return nil, ErrConnectionFailed.Wrap(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrSendFailed.Wrap(err).WithMsg("读取响应失败")
	}

	// 解析响应
	return d.parseResponse(resp.StatusCode, respBody)
}

// parseResponse 解析 Mandrill API 响应
func (d *MandrillDriver) parseResponse(statusCode int, body []byte) (*Result, error) {
	// Mandrill 返回数组格式
	var responses []struct {
		ID           string `json:"_id"`
		Email        string `json:"email"`
		Status       string `json:"status"`
		RejectReason string `json:"reject_reason"`
	}

	if err := json.Unmarshal(body, &responses); err != nil {
		// 可能是错误响应
		var errResp struct {
			Status  string `json:"status"`
			Code    int    `json:"code"`
			Name    string `json:"name"`
			Message string `json:"message"`
		}
		if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil && errResp.Status == "error" {
			return nil, ErrSendFailed.WithMsgf("Mandrill API 错误: %s - %s", errResp.Name, errResp.Message)
		}
		return nil, ErrSendFailed.Wrap(err).WithMsg("解析响应失败")
	}

	if len(responses) == 0 {
		return nil, ErrSendFailed.WithMsg("Mandrill 返回空响应")
	}

	// 取第一个响应
	first := responses[0]

	result := &Result{
		MessageID: first.ID,
		Status:    first.Status,
		Success:   first.Status == "sent" || first.Status == "queued" || first.Status == "scheduled",
	}

	if !result.Success {
		reason := first.RejectReason
		if reason == "" {
			reason = first.Status
		}
		// 返回 result 和 error，让调用方可以获取详情
		return result, ErrSendFailed.WithMsgf("邮件发送失败: %s", reason)
	}

	return result, nil
}

func init() {
	// 注册 Mandrill 驱动到默认注册表
	RegisterDriver(DriverMandrill, NewMandrillDriver)
}
