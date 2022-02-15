package transformer

import "github.com/gofrs/uuid"

type ResourceTypeMetadata struct {
	CycleMetadata
	Type string
}

type ResourceMetadata struct {
	ResourceTypeMetadata
	ResourceId string
}

type CycleMetadata struct {
	CycleId uuid.UUID
}

type RuleResult struct {
	Findings []Finding   `json:"findings"`
	Resource interface{} `json:"resource"`
}

type Finding struct {
	Result interface{} `json:"result"`
	Rule   interface{} `json:"rule"`
}
