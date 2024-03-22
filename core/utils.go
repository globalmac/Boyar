package core

import (
	"bytes"
	"github.com/globalmac/boyar/types"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strings"
	"unicode"
)

type ASTTransformer struct{}

// CreateDir - создание новой директории
func CreateDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

// LoadConfig - загрузка и парсинг файла конфигурации
func LoadConfig(path string) (SiteConfig, error) {

	var config SiteConfig
	configFile, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return config, err
	}

	return config, err
}

// slugify - очистка и подготовка строки для URL
func slugify(s string) string {

	s = strings.ToLower(s)

	var result strings.Builder
	var lastCharDash bool

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			result.WriteRune(r)
			lastCharDash = false
		} else if !lastCharDash {
			result.WriteRune('-')
			lastCharDash = true
		}
	}

	return strings.Trim(result.String(), "-")
}

// slugify - очистка и подготовка строки для URL
func slugifyWithExt(s string) string {
	return s + ".html"
}

// markdownParser - парсинг страницы с разметкой
func markdownParser(content string) (types.MarkdownPost, string) {

	lines := strings.Split(content, "\n")
	count := 1

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if count == 0 && (trimmedLine == "---" || trimmedLine == "") {
			count--
			lines = append(lines[:i], lines[i+1:]...)
			break
		}

		if strings.HasPrefix(trimmedLine, "---") {
			count--
			lines = append(lines[:i], lines[i+1:]...)
			break
		}

		count++
	}

	content = strings.Join(lines, "\n")

	parts := strings.Split(content, "\n---\n")

	var fmd types.MarkdownPost
	var body string
	if len(parts) == 2 {
		body = parts[1]
	}

	err := yaml.Unmarshal([]byte(parts[0]), &fmd)
	if err != nil {
		log.Println(err)
	}

	if fmd.Tags != nil {
		for i, tag := range fmd.Tags {
			fmd.Tags[i] = strings.TrimSpace(tag)
		}
	}

	return fmd, body
}

// sliceContains - поиск строки в слайсе
func sliceContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// markdownRender - рендер Markdown
func markdownRender(markdown string) (string, error) {
	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			extension.Footnote,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithASTTransformers(
				util.Prioritized(&ASTTransformer{}, 10000),
			),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)
	err := md.Convert([]byte(markdown), &buf)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (g *ASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch v := n.(type) {
		case *ast.Link:
			link := v.Destination
			if bytes.HasPrefix(link, []byte("http")) {
				v.SetAttributeString("target", []byte("_blank"))
			}
		}

		return ast.WalkContinue, nil
	})
}

// removeHTMLTags - очистка строки от HTML
func removeHTMLTags(html string) string {
	var result strings.Builder
	var insideTag bool

	for _, char := range html {
		if char == '<' {
			insideTag = true
			continue
		}
		if char == '>' {
			insideTag = false
			continue
		}
		if !insideTag && !unicode.IsControl(char) {
			result.WriteRune(char)
		}
	}

	return result.String()
}

func GetPwd() string {
	p, err := os.Getwd()
	if err != nil {
		return ""
	}
	return p
}
