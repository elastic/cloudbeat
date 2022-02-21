package opa

import (
	"github.com/open-policy-agent/opa/sdk"
	"net/http"
)

type Evaluator struct {
	bundleServer *http.Server
	opa          *sdk.OPA
}

type Metadata struct {
	OpaVersion string `json:"opa_version"`
}

type RuleResult struct {
	Findings []Finding `json:"findings"`
	Metadata Metadata  `json:"metadata"`
	// Golang 1.18 will introduce generics which will be useful for typing the resource field
	Resource interface{} `json:"resource"`
}

type Finding struct {
	Result Result `json:"result"`
	Rule   Rule   `json:"rule"`
}

type Result struct {
	Evaluation string      `json:"evaluation"`
	Evidence   interface{} `json:"evidence"`
}

type Rule struct {
	Benchmark   string   `json:"benchmark"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`
	Name        string   `json:"name"`
	Remediation string   `json:"remediation"`
	Tags        []string `json:"tags"`
}
