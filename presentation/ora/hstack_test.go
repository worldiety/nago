package ora

import (
	"encoding/json"
	"testing"
)

func Test_omit(t *testing.T) {
	var tmp HStack
	buf, err := json.Marshal(tmp)
	t.Log(string(buf), err)
}
