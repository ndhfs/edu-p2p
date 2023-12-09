package dto

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

type Common struct {
	Message *Message `json:"message,omitempty"`
}

func NewTextMessage(from string, text string) Common {
	return Common{
		Message: &Message{
			From: Client{
				Name: from,
			},
			Text: text,
		},
	}
}

func NewMediaMessage(from string, filePath string) (Common, error) {
	// Получаем информацию о файле
	fileStat, err := os.Stat(filePath)
	if err != nil {
		return Common{}, fmt.Errorf("file %s does not exists", filePath)
	}

	// Открываем дескриптор
	src, err := os.Open(filePath)
	if err != nil {
		return Common{}, fmt.Errorf("error read file. %w", err)
	}
	// Обязательно закроем в конце выполнения функции
	defer func() {
		src.Close()
	}()

	// Получим контент сообщения
	fileContent, err := io.ReadAll(src)
	if err != nil {
		return Common{}, fmt.Errorf("error read file content. %w", err)
	}

	// Сформируем структуру
	return Common{
		Message: &Message{
			From: Client{
				Name: from,
			},
			Media: &Media{
				Filename: fileStat.Name(),
				Content:  base64.StdEncoding.EncodeToString(fileContent),
			},
		},
	}, nil
}

type Message struct {
	From  Client `json:"from,omitempty"`
	Text  string `json:"text,omitempty"`
	Media *Media `json:"media,omitempty"`
}

type Media struct {
	Filename string `json:"filename,omitempty"`
	Content  string `json:"content,omitempty"`
}

type Client struct {
	Id   uint   `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
