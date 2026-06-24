package util

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var allowedImageExt = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".bmp":  true,
}

func SaveUploadedImage(file *multipart.FileHeader, dir string, maxBytes int64) (string, error) {
	if file.Size > maxBytes {
		return "", errors.New("文件过大")
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExt[ext] {
		return "", errors.New("不支持的图片格式")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := uuid.NewString() + ext
	dst := filepath.Join(dir, name)

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return "", err
	}
	// 返回 URL 风格的路径
	return "/" + filepath.ToSlash(dst), nil
}

func EnsureDir(p string) error {
	return os.MkdirAll(p, 0o755)
}
