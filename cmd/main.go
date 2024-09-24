package main

import (
	"fmt"
	"k8s-java-thread-dumper/global"
	"k8s-java-thread-dumper/internal/app/handler/grafana"
	"k8s-java-thread-dumper/internal/app/handler/prometheus"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Grafana handler
	grafanaHandler, err := grafana.NewAlertHookHandler()
	if err != nil {
		panic(err)
	}

	// Prometheus handler
	prometheusHandler, err := prometheus.NewAlertHookHandler()
	if err != nil {
		panic(err)
	}

	router := gin.Default()

	router.POST("/hooks/grafana", grafanaHandler)
	router.POST("/hooks/prometheus", prometheusHandler)
	router.StaticFS("/stacks", http.Dir("stacks"))

	port := global.NOTIFY_VIPER.GetInt("server.port")
	if port == 0 {
		port = 8080 // 默认端口
	}

	err = router.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
}
