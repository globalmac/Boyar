package cmd

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/globalmac/boyar/core"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func Serve(cf string) {

	Build(cf)

	config, err := core.LoadConfig("./" + cf)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/", http.FileServer(http.Dir(config.BuildDir)))

		server := http.Server{
			Addr:    ":" + config.Port,
			Handler: mux,
		}
		fmt.Println("-------")
		fmt.Println("Файл конфигурации - ", cf)
		fmt.Println("-------")
		fmt.Println("Запущен на http://localhost:" + config.Port)
		log.Fatal(server.ListenAndServe())
	}()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Ошибка создания watcher:", err)
		return
	}
	defer watcher.Close()

	processEvents := func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					Build(cf)
					fmt.Println("Изменен:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("Ошибка:", err)
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		processEvents()
	}()

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Ошибка при обходе файлов:", err)
			return nil
		}
		if info.Mode().IsDir() || info.IsDir() {
			if strings.HasPrefix(path, "app/design") || strings.HasPrefix(path, "content") {
				return watcher.Add(path)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println("Ошибка при начальном добавлении файлов в watcher:", err)
		return
	}

	done := make(chan bool)
	<-done

	wg.Wait()

}
