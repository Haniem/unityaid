package forms

type Form struct {
	ID     *string     `json:"id"`
	Lang   string      `json:"lang"`
	Fields []FormField `json:"fields"`
	Meta   FormMeta    `json:"meta"`
}

type FormMeta struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type FormField struct {
	Name           string          `json:"name,omitempty"`
	Code           string          `json:"code"`
	Type           string          `json:"type,omitempty"`
	Require        bool            `json:"require"`
	Disabled       bool            `json:"disabled,omitempty"`
	Value          any             `json:"value"`
	MaxLength      int             `json:"maxLength,omitempty"`
	Min            *int            `json:"min,omitempty"`
	Rows           int             `json:"rows,omitempty"`
	Placeholder    string          `json:"placeholder,omitempty"`
	Help           string          `json:"help,omitempty"`
	Endpoint       string          `json:"endpoint,omitempty"`
	UploadEndpoint string          `json:"uploadEndpoint,omitempty"`
	Accept         string          `json:"accept,omitempty"`
	Multi          bool            `json:"multi,omitempty"`
	PossibleValues []PossibleValue `json:"possibleValues,omitempty"`
}

type PossibleValue struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Avatar *string `json:"avatar,omitempty"`
}
