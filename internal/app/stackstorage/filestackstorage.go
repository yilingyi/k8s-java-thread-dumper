package stackstorage

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"k8s-java-thread-dumper/global"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileStackStorage struct {
}

func NewFileStackStorage() FileStackStorage {
	return FileStackStorage{}
}

// markdown消息体构建
func CreateMarkdownMessage(applicationName string, environmentName string, podName string, nodeName string, currentTime string, arthasFileURL string) (string, error) {
	// 创建消息内容
	content := fmt.Sprintf(
		"##### 容器负载高，请尽快处理\n>**应用:** %s\n>**环境:** %s\n>**容器:** %s\n>**节点:** %s\n>**时间:** %s\n>**arthas文件:** [点击查看](%s)",
		applicationName, environmentName, podName, nodeName, currentTime, arthasFileURL)

	// 创建消息数据
	data := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": content,
		},
	}

	// 转换数据为 JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("json marshal error: %v", err)
	}

	return string(jsonData), nil
}

func (s FileStackStorage) Store(model ContainerStackModel) error {
	now := time.Now()
	currentTime := now.Format("2006-01-02T150405")
	dirName := filepath.Join("stacks", now.Format("2006-01-02"), model.Namespace)
	domain := global.NOTIFY_VIPER.GetString("server.domain")
	webhookURL := global.NOTIFY_VIPER.GetString("wework.webhook")

	err := os.MkdirAll(dirName, os.ModePerm)

	if err != nil {
		return fmt.Errorf("mkdir '%s', error:%v", dirName, err)
	}

	fileName := fmt.Sprintf("%s-%s-%s.log", model.PodName, model.ContainerName, now.Format("15-04-05"))

	filePath := filepath.Join(dirName, fileName)

	uri := strings.ReplaceAll(filePath, `\`, `/`)

	arthasFileURL := fmt.Sprintf("%s/%s", domain, uri)
	fmt.Println(arthasFileURL)

	err = os.WriteFile(filePath, []byte(model.Stack), 0644)

	if err != nil {
		return fmt.Errorf("write file '%s', error:%v", filePath, err)
	} else {
		// 创建 Markdown 消息
		jsonMessage, err := CreateMarkdownMessage(model.ContainerName, model.Namespace, model.PodName, model.Node, currentTime, arthasFileURL)
		if err != nil {
			return fmt.Errorf("创建 Markdown 消息出错: %v", err)
		}

		// 创建自定义的 HTTP 客户端，忽略 HTTPS 证书验证
		customTransport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: customTransport}

		// 发送 webhook 通知
		resp, err := client.Post(webhookURL, "application/json", bytes.NewBuffer([]byte(jsonMessage)))
		if err != nil {
			return fmt.Errorf("post webhook error: %v", err)
		}
		defer resp.Body.Close()

		// 检查响应状态
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	}

	return nil
}
