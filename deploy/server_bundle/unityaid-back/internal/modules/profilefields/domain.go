package profilefields

import "encoding/json"

type Group struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organizationId"`
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	SortOrder      int     `json:"sortOrder"`
	IsSystem       bool    `json:"isSystem"`
	IsActive       bool    `json:"isActive"`
	Fields         []Field `json:"fields"`
}

type Field struct {
	ID             string   `json:"id"`
	OrganizationID string   `json:"organizationId"`
	GroupID        string   `json:"groupId"`
	Code           string   `json:"code"`
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	Required       bool     `json:"required"`
	IsSystem       bool     `json:"isSystem"`
	IsActive       bool     `json:"isActive"`
	EditableByUser bool     `json:"editableByUser"`
	SortOrder      int      `json:"sortOrder"`
	Placeholder    string   `json:"placeholder"`
	Help           string   `json:"help"`
	Options        []Option `json:"options"`
}

type Option struct {
	ID        string `json:"id"`
	Value     string `json:"value"`
	Label     string `json:"label"`
	SortOrder int    `json:"sortOrder"`
	IsActive  bool   `json:"isActive"`
}

type SchemaResponse struct {
	Items []Group `json:"items"`
}

type ValuesResponse struct {
	Values map[string]json.RawMessage `json:"values"`
}

type GroupRequest struct {
	OrganizationID string `json:"organizationId" binding:"required"`
	Name           string `json:"name" binding:"required,min=2"`
	Description    string `json:"description"`
	SortOrder      int    `json:"sortOrder"`
	IsActive       bool   `json:"isActive"`
}

type FieldRequest struct {
	OrganizationID string   `json:"organizationId" binding:"required"`
	GroupID        string   `json:"groupId" binding:"required"`
	Name           string   `json:"name" binding:"required,min=2"`
	Type           string   `json:"type" binding:"required"`
	Required       bool     `json:"required"`
	EditableByUser bool     `json:"editableByUser"`
	IsActive       bool     `json:"isActive"`
	SortOrder      int      `json:"sortOrder"`
	Placeholder    string   `json:"placeholder"`
	Help           string   `json:"help"`
	Options        []Option `json:"options"`
}

type ValuesRequest struct {
	OrganizationID string                     `json:"organizationId" binding:"required"`
	Values         map[string]json.RawMessage `json:"values" binding:"required"`
}
