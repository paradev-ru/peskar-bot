package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	chatId                string
	parseMode             string
	disableWebPagePreview bool
	disableNotification   bool
	url                   string
}

func New(config Config) *Client {
	return &Client{
		chatId:                config.ChatId,
		parseMode:             config.ParseMode,
		disableWebPagePreview: config.DisableWebPagePreview,
		disableNotification:   config.DisableNotification,
		url:                   config.URL + config.Token + "/sendMessage",
	}
}

func (c *Client) action(chatId, parseMode, message string, disableWebPagePreview, disableNotification bool) error {
	if chatId == "" {
		chatId = c.chatId
	}

	if parseMode == "" {
		parseMode = c.parseMode
	}

	if parseMode != "" && parseMode != "Markdown" && parseMode != "HTML" {
		return fmt.Errorf("ParseMode %s is not valid, please use 'Markdown' or 'HTML'", parseMode)
	}

	postData := make(map[string]interface{})
	postData["chat_id"] = chatId
	postData["text"] = message

	if parseMode != "" {
		postData["parse_mode"] = parseMode
	}

	if disableWebPagePreview || c.disableWebPagePreview {
		postData["disable_web_page_preview"] = true
	}

	if disableNotification || c.disableNotification {
		postData["disable_notification"] = true
	}

	var post bytes.Buffer
	enc := json.NewEncoder(&post)
	err := enc.Encode(postData)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.url, "application/json", &post)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		type response struct {
			Description string `json:"description"`
			ErrorCode   int    `json:"error_code"`
			Ok          bool   `json:"ok"`
		}
		res := &response{}

		err = json.Unmarshal(body, res)

		if err != nil {
			return fmt.Errorf("Failed to understand Telegram response (err: %s). url: %s data: %v code: %d content: %s", err.Error(), c.url, &postData, resp.StatusCode, string(body))
		}
		return fmt.Errorf("SendMessage error (%d) description: %s", res.ErrorCode, res.Description)

	}
	return nil
}

func (c *Client) Send(message string) error {
	return c.action(c.chatId, c.parseMode, message, c.disableWebPagePreview, c.disableNotification)
}

func (c *Client) ChatSend(chatId, message string) error {
	return c.action(chatId, c.parseMode, message, c.disableWebPagePreview, c.disableNotification)
}
