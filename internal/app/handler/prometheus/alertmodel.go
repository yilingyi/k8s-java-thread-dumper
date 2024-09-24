package prometheus

import (
	"encoding/json"
	"fmt"
	"log"
)

type PrometheusAlert struct {
	Status string         `json:"status"`
	Alerts []AlertDetails `json:"alerts"`
}

type AlertDetails struct {
	Labels Labels `json:"labels"`
}

type Labels struct {
	Pod       string `json:"pod"`
	Node      string `json:"node"`
	Container string `json:"container"`
	Namespace string `json:"namespace"`
}

func (m PrometheusAlert) IsOk() bool {
	return m.Status == "resolved"
}

func (m PrometheusAlert) IsAlerting() bool {
	return m.Status == "firing"
}

func main() {
	data := `...` // 这里替换为你的 JSON 数据

	var alert PrometheusAlert
	err := json.Unmarshal([]byte(data), &alert)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	fmt.Printf("IsOk: %v\n", alert.IsOk())
	fmt.Printf("IsAlerting: %v\n", alert.IsAlerting())

	for _, alertDetail := range alert.Alerts {
		fmt.Printf("Pod: %s\n", alertDetail.Labels.Pod)
		fmt.Printf("Node: %s\n", alertDetail.Labels.Node)
		fmt.Printf("Container: %s\n", alertDetail.Labels.Container)
		fmt.Printf("Namespace: %s\n", alertDetail.Labels.Namespace)
	}
}
