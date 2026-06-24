package email

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/smtp"
	"sync"
	"time"

	"fool_blog_go/internal/config"
)

type codeEntry struct {
	Code      string
	CreatedAt time.Time
}

var (
	codes   = make(map[string]*codeEntry)
	codesMu sync.Mutex
)

func init() {
	go func() {
		for {
			time.Sleep(time.Minute)
			codesMu.Lock()
			now := time.Now()
			for k, v := range codes {
				if now.Sub(v.CreatedAt) > 5*time.Minute {
					delete(codes, k)
				}
			}
			codesMu.Unlock()
		}
	}()
}

func genCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func SendVerifyCode(emailAddr string) (string, error) {
	cfg := config.Current
	if cfg == nil || cfg.Email.Host == "" {
		return "", fmt.Errorf("邮件服务未配置")
	}

	code := genCode()

	codesMu.Lock()
	codes[emailAddr] = &codeEntry{Code: code, CreatedAt: time.Now()}
	codesMu.Unlock()

	addr := fmt.Sprintf("%s:%d", cfg.Email.Host, cfg.Email.Port)
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: Fool Blog 注册验证码\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n您的注册验证码：%s\r\n有效期 5 分钟，请勿泄露给他人。\r\n",
		cfg.Email.From, emailAddr, code))

	// 465 端口用 SSL 直连
	tlsConn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return "", fmt.Errorf("SSL连接失败: %v", err)
	}
	defer tlsConn.Close()

	client, err := smtp.NewClient(tlsConn, cfg.Email.Host)
	if err != nil {
		return "", fmt.Errorf("SMTP客户端创建失败: %v", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", cfg.Email.User, cfg.Email.Password, cfg.Email.Host)
	if err := client.Auth(auth); err != nil {
		return "", fmt.Errorf("SMTP认证失败: %v", err)
	}
	if err := client.Mail(cfg.Email.From); err != nil {
		return "", fmt.Errorf("发件人设置失败: %v", err)
	}
	if err := client.Rcpt(emailAddr); err != nil {
		return "", fmt.Errorf("收件人设置失败: %v", err)
	}
	w, err := client.Data()
	if err != nil {
		return "", fmt.Errorf("数据通道打开失败: %v", err)
	}
	if _, err := w.Write(msg); err != nil {
		return "", fmt.Errorf("邮件写入失败: %v", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("邮件发送失败: %v", err)
	}

	return code, nil
}

func VerifyCode(emailAddr, code string) bool {
	codesMu.Lock()
	defer codesMu.Unlock()
	entry, ok := codes[emailAddr]
	if !ok {
		return false
	}
	if time.Since(entry.CreatedAt) > 5*time.Minute {
		delete(codes, emailAddr)
		return false
	}
	if entry.Code != code {
		return false
	}
	delete(codes, emailAddr)
	return true
}
