package service

import (
	"bytes"
	"regexp"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var md goldmark.Markdown

func init() {
	md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			extension.TaskList,
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(false),
					chromahtml.WithLineNumbers(false),
				),
			),
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
}

func RenderMarkdown(src string) (string, error) {
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type TOCItem struct {
	Level int
	Title string
	ID    string
}

// 从渲染后的 HTML 中抓 H1-H4 标题及其真实锚点 ID。
var htmlHeadingRe = regexp.MustCompile(`(?is)<h([1-4])\s+id="([^"]+)"[^>]*>(.*?)</h[1-4]>`)
var tagStripRe = regexp.MustCompile(`<[^>]+>`)

func ExtractTOC(renderedHTML string) []TOCItem {
	var toc []TOCItem
	for _, m := range htmlHeadingRe.FindAllStringSubmatch(renderedHTML, -1) {
		level := int(m[1][0] - '0')
		id := m[2]
		title := strings.TrimSpace(tagStripRe.ReplaceAllString(m[3], ""))
		if title == "" {
			continue
		}
		toc = append(toc, TOCItem{Level: level, Title: title, ID: id})
	}
	return toc
}

// MakeSummary 从 Markdown 中取摘要。
var headingMDRe = regexp.MustCompile(`(?im)^#{1,6}\s+.*$`)

func MakeSummary(src string, max int) string {
	s := headingMDRe.ReplaceAllString(src, "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}
