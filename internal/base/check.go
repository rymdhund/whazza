package base

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Check struct {
	CheckType   string                 `json:"type"`
	Namespace   string                 `json:"namespace"`
	CheckParams map[string]interface{} `json:"params"`
	Interval    int                    `json:"interval"`
}

func (chk Check) Key() string {
	return fmt.Sprintf("%+v chk", chk)
}

func (chk Check) Name() string {
	return fmt.Sprintf("%s %s", chk.Namespace, chk.CheckType)
}

func (chk Check) ParamsEncoded() []byte {
	params := new(bytes.Buffer)
	json.NewEncoder(params).Encode(chk.CheckParams)
	return params.Bytes()
}

func DecodeParams(inp []byte) map[string]interface{} {
	var ret map[string]interface{}
	json.NewDecoder(bytes.NewReader(inp)).Decode(&ret)
	return ret
}
