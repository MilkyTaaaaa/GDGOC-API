package gemini

// Request untuk rekomendasi
type RecommendationReq struct{
	Query	string	`json:"query"`
	MaxPrice	float64	`json:"max_price,omitempty"`
	Diet	string	`json:"diet,omitempty"`
	Exclude []string	`json:"exclude,omitempty"`
}

// Hasil yang dikembalikan
type RecommendationResult struct{
	Query	string		`json:"query"`
	Recommendations	[]MenuRecommendation	`json:"recommendations"`
	SearchSummary	string	`json:"search_summary"`
	Suggestions	[]string	`json:"suggestions,omitempty"`
}

// rekomendasi per menu
type MenuRecommendation	struct{
	Menu	interface{}	`json:"menu"`
	MatchReason	string	`json:"match_reason`
} 