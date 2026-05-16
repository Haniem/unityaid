package admin

type EntityConfig struct {
	Code        string   `json:"code"`
	Label       string   `json:"label"`
	Table       string   `json:"-"`
	PrimaryKey  string   `json:"primaryKey"`
	OrderBy     string   `json:"-"`
	Columns     []string `json:"columns"`
	Editable    []string `json:"editable"`
	Search      []string `json:"-"`
	CanCreate   bool     `json:"canCreate"`
	CanDelete   bool     `json:"canDelete"`
	Description string   `json:"description"`
}

type ListResponse struct {
	Items []map[string]any `json:"items"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

type EntitiesResponse struct {
	Items []EntityConfig `json:"items"`
}

type UpsertRequest struct {
	Data map[string]any `json:"data" binding:"required"`
}
