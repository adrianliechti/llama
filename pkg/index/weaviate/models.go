package weaviate

type Object struct {
	ID string `json:"id"`

	Created int64 `json:"creationTimeUnix"`
	Updated int64 `json:"lastUpdateTimeUnix"`

	Properties map[string]string `json:"properties"`
}
