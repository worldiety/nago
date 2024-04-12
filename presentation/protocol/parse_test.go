package protocol

import (
	"encoding/json"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	var cfg any
	cfg = ConfigurationRequested{}
	err := json.Unmarshal([]byte(` {"type":"NewConfigurationRequested","requestId":2,"acceptLanguage":"de","colorScheme":"light"}`), &cfg)
	if err != nil {
		t.Fatal(err)
	}
}
