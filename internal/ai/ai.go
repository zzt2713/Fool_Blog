package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"fool_blog_go/internal/config"
)

// ReviewStream sends article content to AI and returns a channel of text chunks.
// Each string sent on the channel is a small piece of the AI response.
// The channel is closed when the response is complete or an error occurs.
func ReviewStream(articleTitle, articleContent string) (<-chan string, error) {
	cfg := config.Current
	if cfg == nil || cfg.AI.APIKey == "" {
		return nil, fmt.Errorf("AI API key not configured")
	}

	baseURL := strings.TrimRight(cfg.AI.BaseURL, "/")
	url := baseURL + "/chat/completions"

	prompt := fmt.Sprintf(`你是一位专业的文章点评助手。请对以下文章进行详细的点评，包括：
1. 文章的核心主题和主要内容
2. 写作亮点和值得学习的地方
3. 可以改进的地方和建议

要求：回复200字以内，简洁精炼。使用 Markdown 格式（如加粗、列表等），写出来。核心主题、写作亮点、改进建议，三个模块，不能换行，不能使用井号的标题格式，只能加粗例如：
**核心主题**：xxxxxxxxxx
**写作亮点**：xxxxxxxxxx
**改进建议**：xxxxxxxxxx
这样写，
文章标题：%s

文章内容（截取）：
%s`,
		truncate(articleTitle, 100),
		truncate(stripHTML(articleContent), 5000),
	)

	body := map[string]interface{}{
		"model": cfg.AI.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"stream": true,
		"max_tokens": 1024,
		"temperature": 0.7,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.AI.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AI request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("AI API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	ch := make(chan string, 32)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				ch <- chunk.Choices[0].Delta.Content
			}
		}
	}()

	return ch, nil
}

// stripHTML removes HTML tags from a string.
func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// truncate shortens a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Review generates a complete AI review (non-streaming) and returns the full text.
func Review(articleTitle, articleContent string) (string, error) {
	cfg := config.Current
	if cfg == nil || cfg.AI.APIKey == "" {
		return "", fmt.Errorf("AI API key not configured")
	}

	baseURL := strings.TrimRight(cfg.AI.BaseURL, "/")
	url := baseURL + "/chat/completions"

	prompt := fmt.Sprintf(`你是一位专业的文章点评助手。请对以下文章进行详细的点评，包括：
1. 文章的核心主题和主要内容
2. 写作亮点和值得学习的地方
3. 可以改进的地方和建议

要求：回复200字以内，简洁精炼。使用 Markdown 格式（如加粗、列表等），写出来。核心主题、写作亮点、改进建议，三个模块，不能换行，不能使用井号的标题格式，只能加粗例如：
**核心主题**：xxxxxxxxxx
**写作亮点**：xxxxxxxxxx
**改进建议**：xxxxxxxxxx
这样写，

文章标题：%s

文章内容（截取）：
%s`,
		truncate(articleTitle, 100),
		truncate(stripHTML(articleContent), 5000),
	)

	body := map[string]interface{}{
		"model": cfg.AI.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 1024,
		"temperature": 0.7,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.AI.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("AI request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("AI API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("AI returned empty response")
	}
	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}
