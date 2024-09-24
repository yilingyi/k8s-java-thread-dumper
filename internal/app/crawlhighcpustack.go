package app

import (
	"bytes"
	"fmt"
	"k8s-java-thread-dumper/global"
	"k8s-java-thread-dumper/internal/util"
	"log"
	"os"
	"strings"
)

type CrawlContext struct {
	Namespace     string
	PodName       string
	ContainerName string
	Node          string
}

const shellFile = "crawl.sh"
const targetFile = "/tmp/crawl.sh"

// CrawlString 函数执行 Arthas 指令，并返回结果
func CrawlString(client util.KubernetesClient, context CrawlContext) (string, error) {
	log.Println("Starting CrawlString function")
	data, err := crawl(client, context)

	if err != nil {
		log.Printf("Error in crawl: %v\n", err)
		return "", err
	}

	str := string(data)

	const startKey = "| plaintext"
	startIndex := strings.Index(str, startKey) + len(startKey) + 2

	str = str[startIndex:]

	const endKey = "[arthas@"
	endIndex := strings.LastIndex(str, endKey) - 2
	str = str[:endIndex]

	log.Println("CrawlString function completed successfully")
	return str, nil
}

// crawl 函数执行 Arthas 指令，并返回原始结果
func crawl(client util.KubernetesClient, context CrawlContext) (stdoutBytes []byte, err error) {
	log.Println("Starting crawl function")
	namespace, podName, containerName, node := context.Namespace, context.PodName, context.ContainerName, context.Node
	log.Printf("Namespace: %s, PodName: %s, ContainerName: %s, Node: %s\n", namespace, podName, containerName, node)
	_ = node

	// Check if arthas download is enabled
	if global.NOTIFY_VIPER.GetBool("arthas.remoteCopy") {
		// Read the arthas-boot.jar file
		arthasJarPath := global.NOTIFY_VIPER.GetString("arthas.path")
		arthasData, err := os.ReadFile(arthasJarPath)
		if err != nil {
			log.Printf("Error reading file %s: %v\n", arthasJarPath, err)
			return nil, fmt.Errorf("read file error: %v", err)
		}

		// Copy arthas-boot.jar to /tmp/arthas in the container
		copyCommands := []string{"/bin/bash", "-c", "mkdir -p /tmp/arthas && cat > /tmp/arthas/arthas-boot.jar"}
		stdin := bytes.NewReader(arthasData)
		log.Println("Copying arthas-boot.jar to Kubernetes pod")

		var copyStdout bytes.Buffer
		stderr, err := client.Exec(namespace, podName, containerName, copyCommands, stdin, &copyStdout)
		if err != nil {
			log.Printf("Error copying arthas-boot.jar: %v\n", err)
			return nil, fmt.Errorf("copy arthas-boot.jar error: %v", err)
		}

		if len(stderr) > 0 && !strings.HasPrefix(string(stderr), "Picked up JAVA_TOOL_OPTIONS:") {
			log.Printf("STDERR: %s\n", string(stderr))
			return nil, fmt.Errorf("STDERR: " + string(stderr))
		}

		// Print the stdout content
		log.Printf("Copy arthas-boot.jar to Kubernetes pod Successfully")
		log.Printf("STDOUT: %s\n", copyStdout.String())
	}

	// executing the shell script
	commands := []string{"/bin/bash", "-c"}
	commands = append(commands, fmt.Sprintf("cp -f /dev/stdin %[1]s;chmod +x %[1]s;%[1]s", targetFile))

	scriptData, err := os.ReadFile(shellFile)
	if err != nil {
		log.Printf("Error reading file %s: %v\n", shellFile, err)
		return nil, fmt.Errorf("read file error: %v", err)
	}

	stdin := bytes.NewReader(scriptData)
	var stdout bytes.Buffer
	log.Println("Executing command in Kubernetes pod")
	stderr, err := client.Exec(namespace, podName, containerName, commands, stdin, &stdout)

	if len(stderr) != 0 && !strings.HasPrefix(string(stderr), "Picked up JAVA_TOOL_OPTIONS:") {
		log.Printf("STDERR: %s\n", string(stderr))
		return nil, fmt.Errorf("STDERR: " + string(stderr))
	}

	if err != nil {
		log.Printf("Error executing command: %v\n", err)
		return nil, err
	}

	log.Println("Command executed successfully")
	return stdout.Bytes(), nil
}
