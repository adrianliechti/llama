package azure

type Results struct {
	Value []Result `json:"value"`
}

type Result map[string]any

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

func (r Result) Location() string {
	if val := r.String("source"); val != "" {
		return val
	}

	return ""
}
