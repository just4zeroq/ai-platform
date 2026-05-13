package model

// Channel upstream configuration for LLM providers.
type Channel struct {
	Id           int    `json:"id"`
	Type         int    `json:"type"`        // provider type: 1=OpenAI, 2=Azure, 3=Custom
	Key          string `json:"key"`         // API key for upstream
	Name         string `json:"name"`        // display name
	Models       string `json:"models"`      // comma-separated model list
	Group        string `json:"group"`       // access group
	Status       int    `json:"status"`      // 1=enabled, 0=disabled
	Priority     int64  `json:"priority"`    // higher = preferred
	Weight       uint   `json:"weight"`      // load balancing weight
	ModelMapping string `json:"model_mapping"` // upstream model name override
	BaseUrl      string `json:"base_url"`    // upstream base URL
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}
