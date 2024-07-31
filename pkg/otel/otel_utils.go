package otel

import (
	"go.opentelemetry.io/otel/attribute"
)

type KeyValue = attribute.KeyValue

func String(key string, val string) KeyValue {
	return attribute.String(key, val)
}

func Strings(key string, val []string) KeyValue {
	return attribute.StringSlice(key, val)
}
