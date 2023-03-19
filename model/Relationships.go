package model

type ContainsIngredient struct {
	Unit   string `json:"unit"`
	Amount int64  `json:"amount"`
	// TODO order
	Relationship
	Resource
}

type Relationship struct {
	SourceId *string `json:"source_id"`
	TargetId string  `json:"target_id"`
}
