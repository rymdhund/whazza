package base

import "fmt"

type Check struct {
	CheckType   string
	Context     string
	CheckParams interface{}
	Interval    int
}

func (chk Check) Key() string {
	return fmt.Sprintf("%+v chk", chk)
}

func (chk Check) Name() string {
	return fmt.Sprintf("%s", chk.CheckType)
}
