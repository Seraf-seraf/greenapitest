//go:build js && wasm

package main

import (
	"regexp"
	"strings"
	"syscall/js"
)

var nonDigitsRegex = regexp.MustCompile(`\D+`)

func main() {
	js.Global().Set("goBuildRequest", js.FuncOf(buildRequest))
	select {}
}

func formatChatID(chatID string) string {
	chatID = strings.TrimSpace(chatID)
	if strings.Contains(chatID, "@") {
		return chatID
	}

	digits := nonDigitsRegex.ReplaceAllString(chatID, "")
	if digits == "" {
		return ""
	}

	return digits + "@c.us"
}

func buildRequest(_ js.Value, args []js.Value) any {
	if len(args) < 7 {
		return map[string]any{"error": "invalid arguments for goBuildRequest"}
	}

	idInstance := strings.TrimSpace(args[0].String())
	apiTokenInstance := strings.TrimSpace(args[1].String())
	method := strings.TrimSpace(args[2].String())
	sendMessageChatID := strings.TrimSpace(args[3].String())
	sendMessageText := strings.TrimSpace(args[4].String())
	sendFileChatID := strings.TrimSpace(args[5].String())
	sendFileURL := strings.TrimSpace(args[6].String())

	if idInstance == "" || apiTokenInstance == "" {
		return map[string]any{"error": "Заполните idInstance и ApiTokenInstance"}
	}

	var payload map[string]any

	switch method {
	case "getSettings", "getStateInstance":
		payload = map[string]any{}

	case "sendMessage":
		if sendMessageChatID == "" || sendMessageText == "" {
			return map[string]any{"error": "Для sendMessage заполните chatId и текст"}
		}

		chatID := formatChatID(sendMessageChatID)
		if chatID == "" {
			return map[string]any{"error": "Некорректный chatId для sendMessage"}
		}

		payload = map[string]any{
			"chatId":  chatID,
			"message": sendMessageText,
		}

	case "sendFileByUrl":
		if sendFileChatID == "" || sendFileURL == "" {
			return map[string]any{"error": "Для sendFileByUrl заполните chatId и URL файла"}
		}

		chatID := formatChatID(sendFileChatID)
		if chatID == "" {
			return map[string]any{"error": "Некорректный chatId для sendFileByUrl"}
		}

		payload = map[string]any{
			"chatId":   chatID,
			"urlFile":  sendFileURL,
			"fileName": "file",
			"caption":  "File by URL",
		}

	default:
		return map[string]any{"error": "Неизвестный метод: " + method}
	}

	return map[string]any{
		"idInstance":       idInstance,
		"apiTokenInstance": apiTokenInstance,
		"method":           method,
		"payload":          payload,
	}
}
