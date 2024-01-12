package jsonschema

type DataType string

const (
	DataTypeObject  DataType = "object"
	DataTypeNumber  DataType = "number"
	DataTypeInteger DataType = "integer"
	DataTypeString  DataType = "string"
	DataTypeArray   DataType = "array"
	DataTypeNull    DataType = "null"
	DataTypeBoolean DataType = "boolean"
)

type FunctionDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Parameters any `json:"parameters,omitempty"`
}

type Definition struct {
	Type        DataType              `json:"type,omitempty"`
	Description string                `json:"description,omitempty"`
	Enum        []string              `json:"enum,omitempty"`
	Properties  map[string]Definition `json:"properties,omitempty"`
	Required    []string              `json:"required,omitempty"`
	Items       *Definition           `json:"items,omitempty"`
}
