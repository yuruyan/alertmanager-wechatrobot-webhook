package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/k8stech/alertmanager-wechatrobot-webhook/model"
	"github.com/k8stech/alertmanager-wechatrobot-webhook/transformer"
	"net/http"
	"strings"
)

// Send send markdown message to dingtalk
//func Send(notification model.Notification, defaultRobot string, grafanaURL string, alertDomain string) (err error) {
//	notificationJSON, _ := json.Marshal(notification)
//	fmt.Println("Nova notification JSON:", string(notificationJSON))
//	markdown, robotURL, err := transformer.TransformToMarkdown(notification, grafanaURL, alertDomain)
//	if err != nil {
//		return
//	}
//
//	data, err := json.Marshal(markdown)
//	if err != nil {
//		return
//	}
//
//	var wechatRobotURL string
//
//	if robotURL != "" {
//		wechatRobotURL = robotURL
//	} else {
//		wechatRobotURL = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=" + defaultRobot
//	}
//
//	req, err := http.NewRequest(
//		"POST",
//		wechatRobotURL,
//		bytes.NewBuffer(data))
//
//	if err != nil {
//		return
//	}
//
//	req.Header.Set("Content-Type", "application/json")
//	client := &http.Client{}
//	resp, err := client.Do(req)
//
//	if err != nil {
//		return
//	}
//
//	defer resp.Body.Close()
//	fmt.Println("response Status:", resp.Status)
//	fmt.Println("response Headers:", resp.Header)
//
//	return
//}

const maxWeChatMessageLength = 4096

// Send sends notifications to WeChat.
func Send(notification model.Notification, defaultRobot string, grafanaURL string, alertDomain string) (err error) {
	notificationJSON, _ := json.Marshal(notification)
	fmt.Println("Nova notification JSON:", string(notificationJSON))

	// Transform alert to markdown format
	markdown, robotURL, err := transformer.TransformToMarkdown(notification, grafanaURL, alertDomain)
	if err != nil {
		return
	}

	// Check the length of the generated markdown content
	content := markdown.Markdown.Content
	if len(content) > maxWeChatMessageLength {
		// Split the content into smaller chunks if it exceeds the limit
		chunks := splitContent(content, maxWeChatMessageLength)
		for _, chunk := range chunks {
			err := sendToWeChat(chunk, robotURL, defaultRobot)
			if err != nil {
				return err
			}
		}
	} else {
		// If within the limit, send as usual
		err = sendToWeChat(content, robotURL, defaultRobot)
	}

	return err
}

// sendToWeChat sends the message to the WeChat webhook
func sendToWeChat(content string, robotURL string, defaultRobot string) error {
	// Prepare the markdown data
	markdown := &model.WeChatMarkdown{
		MsgType: "markdown",
		Markdown: &model.Markdown{
			Content: content,
		},
	}

	// Convert markdown to JSON
	data, err := json.Marshal(markdown)
	if err != nil {
		return err
	}

	// Use robotURL if provided, otherwise use defaultRobot
	var wechatRobotURL string
	if robotURL != "" {
		wechatRobotURL = robotURL
	} else {
		wechatRobotURL = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=" + defaultRobot
	}

	// Create the request
	req, err := http.NewRequest(
		"POST",
		wechatRobotURL,
		bytes.NewBuffer(data))

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Print response for debugging
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	return nil
}

// splitContent splits the content into chunks if it exceeds the max length.
func splitContent(content string, maxLen int) []string {
	var chunks []string
	for len(content) > maxLen {
		// Find a safe split point to avoid breaking markdown syntax (like between tags)
		splitIndex := maxLen
		// Try to find a line break before the maxLen
		if idx := strings.LastIndex(content[:maxLen], "\n"); idx != -1 {
			splitIndex = idx + 1
		}
		// Add chunk and continue
		chunks = append(chunks, content[:splitIndex])
		content = content[splitIndex:]
	}
	// Add the final chunk
	chunks = append(chunks, content)
	return chunks
}
