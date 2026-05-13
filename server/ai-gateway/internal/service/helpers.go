package service

import (
	"math"
	"strings"
)

// estimateTokens estimates token count from text (rough: 1 token ≈ 4 chars).
func estimateTokens(text string) int {
	return int(len(text) / 4) + 1
}

// modelPricing maps model prefixes to per-1K-token cost in platform currency units.
var modelPricing = map[string]struct {
	InputPrice  float64
	OutputPrice float64
}{
	"gpt-4o-mini":        {0.00015, 0.0006},
	"gpt-4o":             {0.005, 0.015},
	"gpt-4-turbo":        {0.01, 0.03},
	"gpt-4-32k":          {0.06, 0.12},
	"gpt-4":              {0.03, 0.06},
	"gpt-3.5-turbo":      {0.0015, 0.002},
	"gpt-3.5":            {0.001, 0.002},
	"claude-3-opus":      {0.015, 0.075},
	"claude-3-sonnet":    {0.003, 0.015},
	"claude-3-5":         {0.003, 0.015},
	"claude-3-7":         {0.003, 0.015},
	"claude-3-haiku":     {0.00025, 0.00125},
	"deepseek-chat":      {0.0005, 0.002},
	"deepseek-reasoner":  {0.0005, 0.002},
	"gemini":             {0.0005, 0.0015},
	"mistral":            {0.0005, 0.0015},
	"llama":              {0.0005, 0.0015},
	"default":            {0.01, 0.03},
}

func getPricing(modelName string) (float64, float64) {
	m := strings.ToLower(modelName)
	for prefix, p := range modelPricing {
		if strings.Contains(m, prefix) {
			return p.InputPrice, p.OutputPrice
		}
	}
	def := modelPricing["default"]
	return def.InputPrice, def.OutputPrice
}

// calcQuota computes cost in quota units (1 quota = 0.001 currency unit).
func calcQuota(modelName string, promptTokens, completionTokens int) int64 {
	inputPrice, outputPrice := getPricing(modelName)
	cost := (float64(promptTokens)/1000)*inputPrice + (float64(completionTokens)/1000)*outputPrice
	return int64(math.Ceil(cost * 1000))
}
