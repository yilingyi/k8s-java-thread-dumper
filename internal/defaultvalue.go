package internal

import (
	"fmt"
	"k8s-java-thread-dumper/global"
	"k8s-java-thread-dumper/internal/app/nodelock"
	"k8s-java-thread-dumper/internal/app/stackstorage"
	"k8s-java-thread-dumper/internal/util"
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var kubernetesClient *util.KubernetesClient

func DefaultKubernetesClient() (*util.KubernetesClient, error) {
	if kubernetesClient != nil {
		return kubernetesClient, nil
	}

	config, err := getConfigByInCluster()
	if err != nil {
		// If fail to get in-cluster config, try to get out-of-cluster config
		config, err = getConfigByOutOfCluster()
		if err != nil {
			log.Println("Error initializing Kubernetes client:", err)
			return nil, err // return the error here
		}
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("new k8s client instance error: %v", err)
	}

	kubernetesClient = &util.KubernetesClient{ClientSet: clientSet, Config: config}
	return kubernetesClient, nil
}

func getConfigByInCluster() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getConfigByOutOfCluster() (*rest.Config, error) {
	configFile := filepath.Join(homeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", configFile)
	if err != nil {
		return nil, fmt.Errorf("build config error: %v", err)
	}
	return config, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

var stackStorage stackstorage.StackStorage = stackstorage.NewFileStackStorage()

func DefaultStackStorage() stackstorage.StackStorage {
	return stackStorage
}

var defaultNodeLockManager nodelock.LockManager

func init() {
	// 从配置中获取值
	maxNodeLockManager := global.NOTIFY_VIPER.GetInt("server.maxNodeLockManager")

	// 使用 maxNodeLockManager 初始化 defaultNodeLockManager
	defaultNodeLockManager = nodelock.NewLockManager(uint(maxNodeLockManager))
}

func GetDefaultNodeLockManager() *nodelock.LockManager {
	return &defaultNodeLockManager
}
