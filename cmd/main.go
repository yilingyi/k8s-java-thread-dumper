package main

import (
	"fmt"
	"k8s-java-thread-dumper/global"
	"k8s-java-thread-dumper/internal/app/handler/grafana"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	handler, err := grafana.NewAlertHookHandler()
	if err != nil {
		panic(err)
	}

	router := gin.Default()

	router.POST("/hooks", handler)
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
