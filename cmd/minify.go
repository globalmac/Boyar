package cmd

import (
	"boyar/core"
	"fmt"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	minifyHtml "github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	minifyJson "github.com/tdewolff/minify/v2/json"
	minifySvg "github.com/tdewolff/minify/v2/svg"
	minifyXml "github.com/tdewolff/minify/v2/xml"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func MinifyFiles(cfg string) {

	config, err := core.LoadConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	directory := config.BuildDir

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		fmt.Printf("Папка назначения %s не найдена\n", directory)
		return
	}

	m := minify.New()

	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", minifyHtml.Minify)
	m.Add("text/html", &minifyHtml.Minifier{
		KeepComments:            false,
		KeepConditionalComments: false,
		KeepSpecialComments:     false,
		KeepDefaultAttrVals:     false,
		KeepDocumentTags:        false,
		KeepEndTags:             false,
		KeepQuotes:              true,
		KeepWhitespace:          false,
	})
	m.AddFunc("image/svg+xml", minifySvg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), minifyJson.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), minifyXml.Minify)

	filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Ошибка чтения пути файла %s: %s\n", path, err)
			return nil
		}

		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".css") || strings.HasSuffix(info.Name(), ".html") || strings.HasSuffix(info.Name(), ".js") || strings.HasSuffix(info.Name(), ".json") || strings.HasSuffix(info.Name(), ".svg") || strings.HasSuffix(info.Name(), ".xml")) {
			content, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("Ошибка чтения файла %s: %s\n", path, err)
				return nil
			}

			var mt = info.Name()

			if strings.HasSuffix(info.Name(), ".css") {
				mt = "text/css"
			} else if strings.HasSuffix(info.Name(), ".html") {
				mt = "text/html"
			} else if strings.HasSuffix(info.Name(), ".js") {
				mt = "text/javascript"
			} else if strings.HasSuffix(info.Name(), ".json") {
				mt = "application/json"
			} else if strings.HasSuffix(info.Name(), ".svg") {
				mt = "image/svg+xml"
			} else if strings.HasSuffix(info.Name(), ".xml") {
				mt = "text/xml"
			}

			minifiedContent, err := m.Bytes(mt, content)
			if err != nil {
				fmt.Printf("Ошибка минификации файла %s: %s\n", path, err)
				return nil
			}

			err = os.WriteFile(path, minifiedContent, 0644)
			if err != nil {
				fmt.Printf("Ошибка записи минифицированного файла %s: %s\n", path, err)
				return nil
			}

		}

		return nil
	})

	fmt.Println("-------")
	fmt.Println("Все статические файлы успешно сжаты/минифицированы!")
}
