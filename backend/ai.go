package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AI 配置通过环境变量提供（兼容 OpenAI / 通义 / DeepSeek 等 OpenAI 风格接口）：
//   AI_API_KEY   必填，模型 appkey（切勿提交到版本库）
//   AI_BASE_URL  选填，默认 https://api.openai.com/v1
//   AI_MODEL     选填，默认 gpt-4o-mini
func aiConfig() (key, baseURL, model string) {
	key = os.Getenv("AI_API_KEY")
	baseURL = os.Getenv("AI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	model = os.Getenv("AI_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}
	return
}

// chatEndpoint 根据 baseURL 推断 chat/completions 完整地址
func chatEndpoint(baseURL string) string {
	if strings.Contains(baseURL, "/chat/completions") {
		return baseURL
	}
	if strings.HasSuffix(baseURL, "/v1") {
		return baseURL + "/chat/completions"
	}
	return baseURL + "/v1/chat/completions"
}

// callChat 调用大模型并返回首条回复文本
func callChat(messages []chatMessage) (string, error) {
	key, baseURL, model := aiConfig()
	if key == "" {
		return "", fmt.Errorf("尚未配置 AI 模型，请设置环境变量 AI_API_KEY")
	}
	reqBody, _ := json.Marshal(chatRequest{Model: model, Temperature: 0.7, Messages: messages})

	httpReq, err := http.NewRequest(http.MethodPost, chatEndpoint(baseURL), bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+key)

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("调用 AI 服务失败: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	var out chatResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("解析 AI 响应失败: %s", string(raw))
	}
	if out.Error != nil {
		return "", fmt.Errorf("AI 服务返回错误: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("AI 未返回内容")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}

type aiCookingReq struct {
	Name        string   `json:"name"`
	Kind        string   `json:"kind"` // dish/soup
	Ingredients []string `json:"ingredients"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// aiCookingMethod 根据菜名与原材料，调用大模型生成烹饪方法
func aiCookingMethod(c *gin.Context) {
	var in aiCookingReq
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	if strings.TrimSpace(in.Name) == "" {
		fail(c, http.StatusBadRequest, "请先填写名称")
		return
	}

	kindLabel := "菜品"
	if in.Kind == KindSoup {
		kindLabel = "汤品"
	}
	ingText := "（未提供原材料）"
	if len(in.Ingredients) > 0 {
		ingText = strings.Join(in.Ingredients, "、")
	}

	prompt := fmt.Sprintf(
		"你是一位经验丰富的中餐厨师。请为%s「%s」给出简洁清晰的烹饪方法。\n已有原材料：%s。\n"+
			"要求：分步骤编号说明，控制在 6 步以内，语言简练实用，只输出步骤本身，不要额外解释。",
		kindLabel, in.Name, ingText,
	)

	content, err := callChat([]chatMessage{
		{Role: "system", Content: "你是专业中餐厨师，擅长用简洁步骤描述菜谱。"},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		fail(c, http.StatusBadGateway, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"cookingMethod": content})
}

type aiFromURLReq struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

var htmlTagRe = regexp.MustCompile(`(?s)<(script|style)[^>]*>.*?</(script|style)>`)
var tagRe = regexp.MustCompile(`(?s)<[^>]+>`)
var wsRe = regexp.MustCompile(`[ \t\r\f\v]+`)
var blankLineRe = regexp.MustCompile(`\n{3,}`)

// aiCookingMethodFromURL 抓取来源网站页面，结合大模型提炼烹饪方法
func aiCookingMethodFromURL(c *gin.Context) {
	var in aiFromURLReq
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	in.URL = strings.TrimSpace(in.URL)
	if in.URL == "" {
		fail(c, http.StatusBadRequest, "请填写来源网址")
		return
	}
	if !strings.HasPrefix(in.URL, "http://") && !strings.HasPrefix(in.URL, "https://") {
		in.URL = "https://" + in.URL
	}

	req, _ := http.NewRequest(http.MethodGet, in.URL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; FoodOrdering/1.0)")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fail(c, http.StatusBadGateway, "抓取网址失败: "+err.Error())
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 最多读 1MB

	// 粗略提取正文文本
	text := htmlTagRe.ReplaceAllString(string(body), " ")
	text = tagRe.ReplaceAllString(text, " ")
	text = html.UnescapeString(text)
	text = wsRe.ReplaceAllString(text, " ")
	text = blankLineRe.ReplaceAllString(text, "\n\n")
	text = strings.TrimSpace(text)
	if len(text) > 8000 {
		text = text[:8000]
	}
	if text == "" {
		fail(c, http.StatusBadGateway, "未能从该网址提取到内容")
		return
	}

	nameHint := ""
	if strings.TrimSpace(in.Name) != "" {
		nameHint = fmt.Sprintf("目标菜品名称：「%s」。\n", in.Name)
	}
	prompt := fmt.Sprintf(
		"%s下面是某菜谱网页的正文内容，请从中提取并整理出该菜品的烹饪方法。\n"+
			"要求：分步骤编号说明，控制在 8 步以内，语言简练实用，只输出烹饪步骤本身，不要额外说明、广告或评论。\n\n网页内容：\n%s",
		nameHint, text,
	)

	content, err := callChat([]chatMessage{
		{Role: "system", Content: "你擅长从杂乱的网页文本中提炼出清晰的菜谱步骤。"},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		fail(c, http.StatusBadGateway, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"cookingMethod": content})
}
