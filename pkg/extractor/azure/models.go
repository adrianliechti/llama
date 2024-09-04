package azure

type OperationStatus string

const (
	OperationStatusSucceeded  OperationStatus = "succeeded"
	OperationStatusRunning    OperationStatus = "running"
	OperationStatusNotStarted OperationStatus = "notStarted"
)

type AnalyzeOperation struct {
	Status OperationStatus `json:"status"`

	Result AnalyzeResult `json:"analyzeResult"`
}

type AnalyzeResult struct {
	ModelID string `json:"modelId"`

	Content string `json:"content"`
}
