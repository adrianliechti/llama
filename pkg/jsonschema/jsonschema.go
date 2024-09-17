package jsonschema

import (
	"encoding/json"
)

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

type Definition struct {
	Type        DataType              `json:"type"`
	Description string                `json:"description,omitempty"`
	Enum        []string              `json:"enum,omitempty"`
	Properties  map[string]Definition `json:"properties"`
	Required    []string              `json:"required,omitempty"`
	Items       *Definition           `json:"items,omitempty"`
}

func (d *Definition) MarshalJSON() ([]byte, error) {
	type Alias Definition

	if d.Type == "" {
		d.Type = DataTypeObject
	}

	if d.Properties == nil {
		d.Properties = make(map[string]Definition)
	}

	return json.Marshal(struct {
		Alias
	}{
		Alias: (Alias)(*d),
	})
}

func (d *Definition) UnmarshalJSON(data []byte) error {
	type Alias Definition

	val := &struct {
		Alias
	}{
		Alias: (Alias)(*d),
	}

	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}

	if val.Alias.Type == "" {
		val.Alias.Type = DataTypeObject
	}

	if val.Alias.Properties == nil {
		val.Alias.Properties = map[string]Definition{}
	}

	*d = Definition(val.Alias)

	return nil
}
