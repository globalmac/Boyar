# Боярин (boyar)

Минималистический статический генератор для простого сайта, блога или лэндинга написанный на GoLang.

Особенности:

- Компактный и расширяемый (можно допиливать как угодно)
- Встроенный постраничный пагинатор для страниц/постов/статей/новостей и т.д.
- Формирование тэгов и категорий (в виде папок структуры контента)
- Отслеживание изменений в контенте (content) и шаблоне (design) через fsnotify
- Встроенный механизм деплоя через SFTP
- Встроенный минификатор HTML/CSS/JS/SVG/JSON/XML
- Генерация Sitemap XML
- Генерация Json файла для поиска по заголовкам

### Установка:

Через Go Get в свой проект

```
go get github.com/globalmac/boyar
```

### Использование:

Перед использование нужно создать в корне проекта папку HTML-шаблона - "**source**" и папку с контентом - "**content**"

```go
package main

import (
	"flag"
	"fmt"
	"github.com/globalmac/boyar/cmd"
)

func main() {

	flag.Parse()

	if flag.NArg() == 0 {
		cmd.Build("config.yaml")
		return
	}

	var command = flag.Arg(0)
	var cnf = flag.Arg(1)

	if cnf == "" {
		cnf = "config.yaml"
	}

	switch command {
	case "build": // Сборка сайта
		cmd.Build(cnf)
	case "serve": // Сборка и превью локально
		cmd.Serve(cnf)
	case "deploy": // SFTP деплой
		cmd.DeployViaSftp(cnf)
	case "min": // Минификация
		cmd.MinifyFiles(cnf)
	case "new": // Создание нового поста/страницы
		cmd.CreateNewPost(flag.Arg(1))
	default:
		fmt.Println("Неизвестная команда:", command)
	}
}

```

TO-DO:

Сделать максимальную расширяемость функционала