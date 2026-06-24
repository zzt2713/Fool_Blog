package service

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ImportResult 描述 Markdown 导入的结果。
type ImportResult struct {
	Title   string
	Content string
	Notes   []string
}

// ImportMarkdown 读取上传的 markdown 文件，处理本地图片，返回新的正文内容。
//   - srcPath: 上传文件在磁盘上的临时位置
//   - filename: 原文件名（用于推断标题）
//   - articleImgDir: 文章图片存放目录（如 uploads/article）
func ImportMarkdown(srcPath, filename, articleImgDir string) (*ImportResult, error) {
	raw, err := os.ReadFile(srcPath)
	if err != nil {
		return nil, err
	}
	content := string(raw)
	title := titleFromMarkdown(content)
	if title == "" {
		title = strings.TrimSuffix(filename, filepath.Ext(filename))
	}

	notes := []string{}
	imgRe := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	content = imgRe.ReplaceAllStringFunc(content, func(match string) string {
		m := imgRe.FindStringSubmatch(match)
		alt, link := m[1], strings.TrimSpace(m[2])
		if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
			return match
		}
		// 本地图片
		base := filepath.Dir(srcPath)
		localPath := link
		if !filepath.IsAbs(localPath) {
			localPath = filepath.Join(base, localPath)
		}
		if _, err := os.Stat(localPath); err != nil {
			notes = append(notes, "图片不存在: "+link)
			return match
		}
		if err := os.MkdirAll(articleImgDir, 0o755); err != nil {
			notes = append(notes, "无法创建图片目录: "+err.Error())
			return match
		}
		ext := filepath.Ext(localPath)
		newName := uuid.NewString() + ext
		dst := filepath.Join(articleImgDir, newName)
		if err := copyFile(localPath, dst); err != nil {
			notes = append(notes, "图片复制失败: "+link)
			return match
		}
		newURL := "/" + filepath.ToSlash(dst)
		return fmt.Sprintf("![%s](%s)", alt, newURL)
	})

	return &ImportResult{Title: title, Content: content, Notes: notes}, nil
}

func titleFromMarkdown(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// ExportSingle 把单篇文章写出 .md 文件，返回磁盘绝对路径与建议文件名。
func ExportSingle(title, content, exportDir string) (string, string, error) {
	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		return "", "", err
	}
	clean := safeFilename(title)
	name := fmt.Sprintf("%s-%s.md", clean, time.Now().Format("20060102"))
	abs := filepath.Join(exportDir, name)
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return "", "", err
	}
	return abs, name, nil
}

// WriteExportZip 将多篇 markdown 文件打包为 zip，写到 http.ResponseWriter。
func WriteExportZip(w http.ResponseWriter, items []struct{ Title, Content string }) error {
	zw := zip.NewWriter(w)
	defer zw.Close()
	for _, it := range items {
		name := safeFilename(it.Title) + ".md"
		f, err := zw.Create(name)
		if err != nil {
			return err
		}
		if _, err := f.Write([]byte(it.Content)); err != nil {
			return err
		}
	}
	return nil
}

var unsafeFileRe = regexp.MustCompile(`[\\/:*?"<>|]`)

func safeFilename(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		s = "untitled"
	}
	s = unsafeFileRe.ReplaceAllString(s, "_")
	if len(s) > 80 {
		runes := []rune(s)
		if len(runes) > 80 {
			s = string(runes[:80])
		}
	}
	return s
}
