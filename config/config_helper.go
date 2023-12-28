package config

import (
	"strings"
)

type modelMapper map[string]modelConfig

func (m modelMapper) From(val string) string {
	for k, v := range m {
		if v.ID != "" && strings.EqualFold(v.ID, val) {
			return k
		}
	}

	for k := range m {
		if strings.EqualFold(k, val) {
			return k
		}
	}

	return ""
}

func (m modelMapper) To(val string) string {
	for k, v := range m {
		if strings.EqualFold(k, val) {
			if v.ID != "" {
				return v.ID
			}

			return k
		}
	}

	return ""
}
