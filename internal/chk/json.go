package chk

import (
	"encoding/json"
	"fmt"
)

func (c Check) MarshalJSON() ([]byte, error) {
	// We need to parse the same json into both the check and the checker
	bs, err := json.Marshal(c.Checker)
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

	c.Checker, err = unmarshalChecker(c.Type, jsonData)
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

func unmarshalChecker(typ string, jsonData []byte) (Checker, error) {
	switch typ {
	case "nop":
		var checker NopChecker
		err := json.Unmarshal(jsonData, &checker)
		if err != nil {
			return nil, fmt.Errorf("Error parsing %s check: %w", typ, err)
		}
		return checker, nil
	case "check-in":
		var checker CheckInChecker
		err := json.Unmarshal(jsonData, &checker)
		if err != nil {
			return nil, fmt.Errorf("Error parsing %s check: %w", typ, err)
		}
		return checker, nil
	case "http-up":
		var checker HttpUpChecker
		err := json.Unmarshal(jsonData, &checker)
		if err != nil {
			return nil, fmt.Errorf("Error parsing %s check: %w", typ, err)
		}
		return checker, nil
	case "https-up":
		var checker HttpsUpChecker
		err := json.Unmarshal(jsonData, &checker)
		if err != nil {
			return nil, fmt.Errorf("Error parsing %s check: %w", typ, err)
		}
		return checker, nil
	case "cert":
		var checker CertChecker
		err := json.Unmarshal(jsonData, &checker)
		if err != nil {
			return nil, fmt.Errorf("Error parsing %s check: %w", typ, err)
		}
		return checker, nil
	default:
		return nil, fmt.Errorf("Unkown Check type: %s", typ)
	}

}
