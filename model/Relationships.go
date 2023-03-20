package model

type ContainsIngredient struct {
	Unit   string `json:"unit"`
	Amount int64  `json:"amount"`
	// TODO order
	Relationship
	Resource
}

func (l ContainsIngredient) Equal(r ContainsIngredient) bool {
	return l.Unit == r.Unit && l.Amount == r.Amount && l.TargetId == r.TargetId
}

type Relationship struct {
	SourceId *string `json:"source_id"`
	TargetId string  `json:"target_id"`
}
