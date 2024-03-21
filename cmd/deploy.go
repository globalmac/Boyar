package cmd

import (
	"fmt"
	"github.com/globalmac/boyar/core"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"path/filepath"
)

func DeployViaSftp(cfg string) {

	appConfig, err := core.LoadConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	config := &ssh.ClientConfig{
		User: appConfig.SftpLogin,
		Auth: []ssh.AuthMethod{
			ssh.Password(appConfig.SftpPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", appConfig.SftpServer+":"+appConfig.SftpPort, config)
	if err != nil {
		log.Fatalf("Ошибка соединения: %v", err)
	}
	defer conn.Close()

	client, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatalf("Ошибка соединения через SFTP клиент: %v", err)
	}
	defer client.Close()

	localFolder := appConfig.BuildDir

	remoteFolder := "/"

	fmt.Println("Начинаем загрузку файлов:")

	err = uploadFolder(client, localFolder, remoteFolder)
	if err != nil {
		log.Fatalf("Ошибка загрузки папки на сервер: %v", err)
	}

	fmt.Println("Папка /dist успешно загружена!")
}

// uploadFolder - Функция для рекурсивной загрузки содержимого папки на сервер
func uploadFolder(client *sftp.Client, localPath, remotePath string) error {
	localFiles, err := os.ReadDir(localPath)
	if err != nil {
		return err
	}

	err = client.MkdirAll(remotePath)
	if err != nil {
		return err
	}

	for _, file := range localFiles {
		localFilePath := filepath.Join(localPath, file.Name())
		remoteFilePath := filepath.Join(remotePath, file.Name())

		if file.IsDir() {
			err = uploadFolder(client, localFilePath, remoteFilePath)
			if err != nil {
				return err
			}
		} else {
			localFile, err := os.Open(localFilePath)
			if err != nil {
				return err
			}
			defer localFile.Close()

			remoteFile, err := client.Create(remoteFilePath)
			if err != nil {
				return err
			}
			defer remoteFile.Close()

			_, err = io.Copy(remoteFile, localFile)
			if err != nil {
				return err
			}
			fmt.Printf("- %s\n", remoteFilePath)
		}
	}

	return nil
}
