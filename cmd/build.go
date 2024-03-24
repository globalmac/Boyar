package cmd

import (
	"fmt"
	"github.com/globalmac/boyar/core"
	"os"
	"text/tabwriter"
	"time"
)

// Build - собираем проект и запускаем сервер локально
func Build(cnf string) {

	start := time.Now()

	var c = core.Process(cnf)
	c.ScanContent()
	c.MakeIndexPage()
	c.MakeTagIndexPage()
	c.MakeDetailPages()
	c.MakeTagPages()
	c.MakeRSS()
	c.MakeSiteMap()
	c.MakeRobotsTxt()
	c.MakeSearchJson()
	c.MakePostCategories()
	c.CopyStaticFiles()

	duration := time.Since(start)

	rows := map[string]int{
		"Страниц/постов": len(c.Posts),
		"Тэгов":          len(c.Tags),
		"Категорий":      len(c.PostTypes),
	}

	w := tabwriter.NewWriter(os.Stdout, 8, 8, 3, '\t', 0)
	fmt.Fprintln(w, "-------")
	fmt.Fprintln(w, "Контент\tИтого")
	for k, v := range rows {
		fmt.Fprintf(w, "%s\t%d\n", k, v)
	}
	fmt.Fprintln(w, "-------")
	fmt.Fprintln(w, "Файл конфигурации - ", cnf)
	fmt.Fprintln(w, "-------")
	fmt.Fprintf(w, "%s\t%s\n", "Время сборки", duration)
	w.Flush()
}
