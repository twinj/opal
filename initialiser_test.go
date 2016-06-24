package opal

import (
	"reflect"
	"testing"
)

func TestExtractOpalTags(t *testing.T) {

	var tagOpalTests = []struct {
		Tag   reflect.StructTag
		Value Tag
	}{
		{`||`, ``},
		{`|something|`, `something`},
		{`json:"value"|with normal|`, `with normal`},
	}
	for _, tt := range tagOpalTests {
		if v := ExtractOpalTags(tt.Tag); v != tt.Value {
			t.Errorf("ExtractOpalTags(%#q) = %#q, want %#q", tt.Tag, v, tt.Value)
		}
	}

}

func TestTagGet(t *testing.T) {
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
	for _, tt := range tagGetTests {
		if v := tt.Tag.Get(tt.Key); v != tt.Value {
			t.Errorf("Tag(%#q).Get(%#q) = %#q, want %#q", tt.Tag, tt.Key, v, tt.Value)
		}
	}
}

func TestImportName(t *testing.T) {
	type T struct{}
	r := reflect.TypeOf(T{})
	s := importName(r)
	if s != "opal.T" {
		t.Errorf("importName(%#q) = %#q, want %#q", r, s, "opal.T")
	}
}
