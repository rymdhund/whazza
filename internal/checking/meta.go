package checking

import (
	"fmt"

	"github.com/rymdhund/whazza/internal/base"
)

type CheckMeta interface {
	ParseParams(base.Check) (interface{}, error)
	DefaultNamespace(base.Check) string
	DoCheck(chk base.Check) base.Result
}

func GetCheckMeta(checkType string) (CheckMeta, error) {
	switch checkType {
	case "http-up":
		return HttpUpCheckMeta{}, nil
	default:
		return nil, fmt.Errorf("Unknown check type: %s", checkType)
	}
}