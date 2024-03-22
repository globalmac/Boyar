package cmd

import (
	"fmt"
	"github.com/globalmac/boyar/core"
	"log"
	"os"
	"path/filepath"
	"time"
)

// CreateNewPost - создаём новую статью по шаблону
func CreateNewPost(path string) {
	if path == "" {
		log.Fatalln("Путь пустой!")
	}

	path = core.GetPwd() + "/content/" + path
	core.CreateDir(filepath.Dir(path))

	markdownContent := fmt.Sprintf("---\ntitle: 111\ndate: %s\ndraft: false\ntags: [\"111\", \"222\"]\nimage: \"/cdn/123/111.png\"\ncover: \"/cdn/123/222.png\"\n---\n\n\n<!--more-->\n\n\n", time.Now().Format("2006-01-02T15:04:05Z"))

	f, err := os.Create(path)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	f.Write([]byte(markdownContent))

	fmt.Printf("Новая страница/пост успешно создана: %s\n", path)
}
