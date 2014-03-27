package opal

import (
	"testing"
)

// Credit to Go
var tagGetTests = []struct {
	Tag   Tag
	Key   string
	Value string
}{
	{`protobuf:"PB(12)"`, `protobuf`, `"PB(12)"`},
	{`protobuf:"PB(12)"`, `foo`, ``},
	{`protobuf:"PB(12)"`, `rotobuf`, ``},
	{`protobuf:"PB(12)" json:"name"`, `json`, `"name"`},
	{`protobuf:"PB(12)" json:"name"`, `protobuf`, `"PB(12)"`},
	{`protobuf: "PB(12)" json: "name"`, `json`, `"name"`},
	{`protobuf: "PB(12)" json: "name"`, `protobuf`, `"PB(12)"`},
	{`protobuf: "PB(12)", json: "name"`, `json`, `"name"`},
	{`protobuf: "PB(12)", json: "name"`, `protobuf`, `"PB(12)"`},
	{`protobuf: "PB(12)",json: "name"`, `json`, `"name"`},
	{`protobuf: "PB(12)",json: "name"`, `protobuf`, `"PB(12)"`},
	{`protobuf:"PB(12)",json:"name"`, `json`, `"name"`},
	{`protobuf:"PB(12)", json:"name"`, `protobuf`, `"PB(12)"`},
}

func TestTagGet(t *testing.T) {
	for _, tt := range tagGetTests {
		if v := tt.Tag.Get(tt.Key); v != tt.Value {
			t.Errorf("Tag(%#q).Get(%#q) = %#q, want %#q", tt.Tag, tt.Key, v, tt.Value)
		}
	}
}
