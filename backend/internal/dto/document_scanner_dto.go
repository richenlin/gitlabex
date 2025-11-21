package dto

// SimpleScanResult 扫描结果
type SimpleScanResult struct {
	TotalDocuments int      `json:"total_documents"`
	NewDocuments   int      `json:"new_documents"`
	UpdatedDocs    int      `json:"updated_docs"`
	Errors         []string `json:"errors,omitempty"`
}
