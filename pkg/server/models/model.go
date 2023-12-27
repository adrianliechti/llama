package models

type Models struct {
	Models []Model `json:"data"`
}

type Model struct {
	ID string `json:"id"`

	Object    string `json:"object"`
	CreatedAt int64  `json:"created"`
}
