package azure

type Results struct {
	Value []Result `json:"value"`
}

type Result map[string]any

func (r Result) ID() string {
	if val := r.String("Id"); val != "" {
		return val
	}

	if val := r.String("id"); val != "" {
		return val
	}

	return ""
}

func (r Result) Title() string {
	if val := r.String("title"); val != "" {
		return val
	}

	return ""
}

func (r Result) Content() string {
	if val := r.String("content"); val != "" {
		return val
	}

	return ""
}

func (r Result) Source() string {
	if val := r.String("source"); val != "" {
		return val
	}

	if val := r.String("location"); val != "" {
		return val
	}

	return ""
}

func (r Result) Metadata() map[string]string {
	if val := r.Map("metadata"); val != nil {
		return val
	}

	return nil
}

func (r Result) String(name string) string {
	val, ok := r[name]

	if !ok {
		return ""
	}

	data, ok := val.(string)

	if !ok {
		return ""
	}

	return data
}

func (r Result) Map(name string) map[string]string {
	val, ok := r[name]

	if !ok {
		return nil
	}

	slice, ok := val.([]interface{})

	if !ok {
		return nil
	}

	if len(slice) == 0 {
		return nil
	}

	result := map[string]string{}

	for _, item := range slice {
		entry, ok := item.(map[string]interface{})

		if !ok {
			continue
		}

		key := entry["key"].(string)
		value := entry["value"].(string)

		if key == "" {
			continue
		}

		result[key] = value
	}

	return result
}
