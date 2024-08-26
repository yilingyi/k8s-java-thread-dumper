package grafana

import (
	"encoding/json"
	"fmt"
	"log"
)

type AlertModel struct {
	State       string           `json:"status"`
	EvalMatches []EvalMatchModel `json:"alerts"`
}

type EvalMatchModel struct {
	Values map[string]float64 `json:"values"`
	Tags   EvalMatchModelTag  `json:"labels"`
}

type EvalMatchModelTag struct {
	Pod       string `json:"pod"`
	Node      string `json:"node"`
	Container string `json:"container"`
	Namespace string `json:"namespace"`
}

func (m AlertModel) IsOk() bool {
	return m.State == "ok"
}

func (m AlertModel) IsAlerting() bool {
	return m.State == "alerting"
}

func main() {
	data := `...` // 这里替换为你的 JSON 数据

	var alert AlertModel
	err := json.Unmarshal([]byte(data), &alert)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	fmt.Printf("IsOk: %v\n", alert.IsOk())
	fmt.Printf("IsAlerting: %v\n", alert.IsAlerting())

	for _, match := range alert.EvalMatches {
		fmt.Printf("Pod: %s\n", match.Tags.Pod)
		fmt.Printf("Container: %s\n", match.Tags.Container)
		for k, v := range match.Values {
			fmt.Printf("Value %s: %f\n", k, v)
		}
	}
}
