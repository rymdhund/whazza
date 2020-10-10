package checking

import (
	"encoding/json"
	"fmt"
)

func (c Check) MarshalJSON() ([]byte, error) {
	// A bit of a hack
	bs, err := json.Marshal(c.Runner)
	if err != nil {
		return nil, err
	}
	mp := map[string]interface{}{}
	mp["interval"] = c.Interval
	mp["namespace"] = c.Namespace
	mp["type"] = c.Type
	err = json.Unmarshal(bs, &mp)
	if err != nil {
		return nil, err
	}

	return json.Marshal(mp)
}

func (c *Check) UnmarshalJSON(jsonData []byte) error {
	var base checkBase
	err := json.Unmarshal(jsonData, &base)
	if err != nil {
		return err
	}
	c.checkBase = base

	c.Runner, err = unmarshalRunner(c.Type, jsonData)
	if err != nil {
		return err
	}

	// Validate that everything looks good
	err = c.Validate()
	if err != nil {
		return err
	}

	return nil
}

func unmarshalRunner(typ string, jsonData []byte) (CheckRunner, error) {
	switch typ {
	case "http-up":
		var runner HttpUpCheck
		err := json.Unmarshal(jsonData, &runner)
		if err != nil {
			return nil, fmt.Errorf("Error parsing http-up check: %w", err)
		}
		return runner, nil
	case "https-up":
		var runner HttpsUpCheck
		err := json.Unmarshal(jsonData, &runner)
		if err != nil {
			return nil, fmt.Errorf("Error parsing http-up check: %w", err)
		}
		return runner, nil
	default:
		return nil, fmt.Errorf("Unkown Check type: %s", typ)
	}

}
