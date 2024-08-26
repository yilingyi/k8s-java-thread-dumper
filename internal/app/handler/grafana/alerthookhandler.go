package grafana

import (
	"encoding/json"
	"fmt"
	"io"
	"k8s-java-thread-dumper/internal"
	"k8s-java-thread-dumper/internal/app"
	"k8s-java-thread-dumper/internal/app/stackstorage"
	"k8s-java-thread-dumper/internal/util"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @summary Kubernetes client instance.
// kubernetesClient保存Kubernetes客户端实例。
var kubernetesClient *util.KubernetesClient

// @summary Storage for stack information.
// stackStorage是存储堆栈信息的存储器。
var stackStorage stackstorage.StackStorage = stackstorage.NewFileStackStorage()

// NewAlertHookHandler creates a new Grafana alert hook handler.
// @description It initializes the Kubernetes client and returns a HookHandler function.
// @tags grafana
// @return 200 {object} gin.Context "Successfully initialized"
// @return 500 {object} gin.Context "Failed to initialize"
// NewAlertHookHandler返回一个可以处理Grafana警报钩子的函数。
// 它初始化Kubernetes客户端并返回HookHandler函数。
func NewAlertHookHandler() (func(ctx *gin.Context), error) {
	client, err := internal.DefaultKubernetesClient()
	if err != nil {
		return nil, fmt.Errorf("create k8s client error: %v", err)
	}

	if client == nil {
		log.Printf("get k8s client return nil: %v\n", err)
		return nil, fmt.Errorf("k8s client is nil")
	}

	kubernetesClient = client
	return HookHandler, nil
}

// HookHandler is the main function to handle Grafana alert hooks.
// @description It reads the request body, processes the alert data, and triggers the alert handler.
// @tags grafana
// @accept  json
// @produce json
// @param data body AlertModel true "Alert Model"
// @param Authorization header string true "Authorization Token"
// @router /api/v1/K8sAlertCallback [post]
// HookHandler是处理Grafana警报钩子的主要函数。
// 它读取请求体，处理警报数据，并触发警报的处理。
func HookHandler(c *gin.Context) {
	log.Println("HookHandler is starting...")
	body := c.Request.Body
	if body == nil {
		c.String(http.StatusBadRequest, "not found body.")
		return
	}
	defer body.Close()

	data, err := io.ReadAll(body)

	if err != nil {
		log.Printf("read all error: %v\n", err)
		c.String(http.StatusBadRequest, "read body error.")
		return
	}

	if len(data) <= 0 {
		c.String(http.StatusBadRequest, "body is empty.")
		return
	}

	var model AlertModel
	fmt.Printf("look here: %+v\n", model)
	err = json.Unmarshal(data, &model)

	if err != nil {
		log.Printf("unmarshal error: %v\n", err)
		c.String(http.StatusBadRequest, "unmarshal error")
		return
	}

	// 使用 fmt.Printf 将接收到的告警信息输出到 stdout
	fmt.Printf("Received alert: %+v\n", model)

	go doHandle(model)

	c.Status(http.StatusAccepted)
	log.Println("HookHandler has finished...")
}

// doHandle processes the alert model and triggers further processing based on the alert condition.
// @description It's a helper function for HookHandler.
// doHandle处理警报模型并根据警报条件触发进一步处理。
func doHandle(model AlertModel) {
	log.Printf("doHandle is starting for model: %+v\n", model)
	// 忽略正常警报
	if model.IsOk() {
		return
	}

	matches := model.EvalMatches
	for _, e := range matches {
		go doHandleEvalMatch(e)
	}

	log.Println("doHandle has finished...")
}

// doHandleEvalMatch handles a single evaluation match in the alert model.
// @description It gets a lock for the associated node, crawls stack information and stores the stack.
// doHandleEvalMatch处理警报模型中的单个评估匹配。
// 它为关联的节点获取锁定，爬取堆栈信息并存储堆栈。
func doHandleEvalMatch(model EvalMatchModel) {
	log.Printf("doHandleEvalMatch is starting for model: %+v\n", model)
	tag := model.Tags
	node := tag.Node

	nodeLockManager := *internal.GetDefaultNodeLockManager()
	fmt.Printf("doHandleEvalMatch: GetLock for node: %s\n", node)
	locker := nodeLockManager.GetLock(node)

	locker.Lock()
	defer locker.Unlock()

	namespace, podName, containerName := tag.Namespace, tag.Pod, tag.Container
	log.Printf("doHandleEvalMatch: Namespace: %s, PodName: %s, ContainerName: %s, Node: %s\n", namespace, podName, containerName, node)
	// 获取arthas信息
	stack, err := app.CrawlString(*kubernetesClient, app.CrawlContext{
		Namespace:     namespace,
		PodName:       podName,
		ContainerName: containerName,
		Node:          node,
	})

	if err != nil {
		log.Printf("crawlString tag: %v, error: %v\n", tag, err)
		return
	}

	err = storeStack(stack, model)

	if err != nil {
		log.Printf("store stack error: %v\n", err)
	}
	log.Printf("doHandleEvalMatch has finished for model: %+v\n", model)
}

// storeStack stores stack information for the given model in the stack storage.
// @description It's a helper function for doHandleEvalMatch.
// storeStack为给定模型在堆栈存储中存储堆栈信息。
func storeStack(stack string, model EvalMatchModel) error {
	tags := model.Tags
	namespace, podName, containerName, node := tags.Namespace, tags.Pod, tags.Container, tags.Node
	err := stackStorage.Store(stackstorage.ContainerStackModel{
		Namespace:     namespace,
		PodName:       podName,
		ContainerName: containerName,
		Node:          node,
		Stack:         stack,
	})

	return err
}
