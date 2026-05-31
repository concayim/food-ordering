package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// NotifyChannel 单个推送渠道
type NotifyChannel interface {
	Name() string
	Enabled() bool
	Send(title, content string) error
}

// NotifyStatus 推送配置状态
type NotifyStatus struct {
	Enabled  bool     `json:"enabled"`
	Channels []string `json:"channels"`
	Message  string   `json:"message"`
}

var notifyChannels []NotifyChannel

func initNotifier() {
	notifyChannels = nil
	if u := os.Getenv("FEISHU_WEBHOOK_URL"); u != "" {
		notifyChannels = append(notifyChannels, &feishuChannel{url: u})
	}
	if u := os.Getenv("WECOM_WEBHOOK_URL"); u != "" {
		notifyChannels = append(notifyChannels, &wecomChannel{url: u})
	}
	if t := os.Getenv("PUSHPLUS_TOKEN"); t != "" {
		notifyChannels = append(notifyChannels, &pushPlusChannel{token: t})
	}

	if len(notifyChannels) == 0 {
		log.Println("订单推送: 未配置（可设置 FEISHU_WEBHOOK_URL / WECOM_WEBHOOK_URL / PUSHPLUS_TOKEN）")
		return
	}
	names := make([]string, len(notifyChannels))
	for i, ch := range notifyChannels {
		names[i] = ch.Name()
	}
	log.Printf("订单推送已启用: %s", strings.Join(names, ", "))
}

func notifierStatus() NotifyStatus {
	channels := make([]string, 0)
	for _, ch := range notifyChannels {
		if ch.Enabled() {
			channels = append(channels, ch.Name())
		}
	}
	if len(channels) == 0 {
		return NotifyStatus{
			Enabled: false,
			Message: "未配置推送渠道。飞书: FEISHU_WEBHOOK_URL；企业微信: WECOM_WEBHOOK_URL；个人微信(PushPlus): PUSHPLUS_TOKEN",
		}
	}
	return NotifyStatus{
		Enabled:  true,
		Channels: channels,
		Message:  "新订单创建后将自动推送到已配置渠道",
	}
}

func formatOrderNotify(order Order) (title, content string) {
	title = fmt.Sprintf("【%s】新订单 #%d", shopName(), order.ID)
	var b strings.Builder
	b.WriteString(fmt.Sprintf("单号: #%d\n", order.ID))
	b.WriteString(fmt.Sprintf("时间: %s\n", order.CreatedAt.Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("状态: %s\n", orderStatusLabel(order.Status)))
	b.WriteString("—— 明细 ——\n")
	total := 0
	for _, it := range order.Items {
		b.WriteString(fmt.Sprintf("· %s × %d\n", it.DishName, it.Quantity))
		total += it.Quantity
	}
	b.WriteString(fmt.Sprintf("合计: %d 件\n", total))
	if strings.TrimSpace(order.Remark) != "" {
		b.WriteString(fmt.Sprintf("备注: %s\n", order.Remark))
	}
	return title, b.String()
}

func notifyOrderAsync(orderID uint) {
	go func() {
		var order Order
		if err := db.Preload("Items").First(&order, orderID).Error; err != nil {
			log.Printf("[notify] 加载订单 #%d 失败: %v", orderID, err)
			return
		}
		notifyOrder(order)
	}()
}

func notifyOrder(order Order) {
	st := notifierStatus()
	if !st.Enabled {
		return
	}
	title, content := formatOrderNotify(order)
	for _, ch := range notifyChannels {
		if !ch.Enabled() {
			continue
		}
		if err := ch.Send(title, content); err != nil {
			log.Printf("[notify] %s 推送失败: %v", ch.Name(), err)
		} else {
			log.Printf("[notify] %s 已推送订单 #%d", ch.Name(), order.ID)
		}
	}
}

func notifyTestMessage() error {
	st := notifierStatus()
	if !st.Enabled {
		return fmt.Errorf("未配置任何推送渠道")
	}
	title := fmt.Sprintf("【%s】推送测试", shopName())
	content := "这是一条测试消息，说明订单推送渠道配置正常。"
	var errs []string
	for _, ch := range notifyChannels {
		if err := ch.Send(title, content); err != nil {
			errs = append(errs, ch.Name()+": "+err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "; "))
	}
	return nil
}

func postJSON(url string, payload any) error {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(raw))
	}
	return nil
}

// ---------- 飞书群机器人 ----------

type feishuChannel struct{ url string }

func (c *feishuChannel) Name() string    { return "feishu" }
func (c *feishuChannel) Enabled() bool   { return c.url != "" }
func (c *feishuChannel) Send(title, content string) error {
	text := title + "\n\n" + content
	return postJSON(c.url, map[string]any{
		"msg_type": "text",
		"content":  map[string]string{"text": text},
	})
}

// ---------- 企业微信群机器人（微信侧常用） ----------

type wecomChannel struct{ url string }

func (c *wecomChannel) Name() string    { return "wecom" }
func (c *wecomChannel) Enabled() bool   { return c.url != "" }
func (c *wecomChannel) Send(title, content string) error {
	text := title + "\n\n" + content
	return postJSON(c.url, map[string]any{
		"msgtype": "text",
		"text":    map[string]string{"content": text},
	})
}

// ---------- PushPlus → 个人微信 ----------

type pushPlusChannel struct{ token string }

func (c *pushPlusChannel) Name() string  { return "pushplus" }
func (c *pushPlusChannel) Enabled() bool { return c.token != "" }
func (c *pushPlusChannel) Send(title, content string) error {
	return postJSON("https://www.pushplus.plus/send", map[string]string{
		"token":   c.token,
		"title":   title,
		"content": strings.ReplaceAll(content, "\n", "<br>"),
	})
}
