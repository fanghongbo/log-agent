package utils

import (
	"regexp"
)

func CheckRegExpr(expr string) bool {
	var (
		err error
	)

	if _, err = regexp.Compile(expr); err != nil {
		return false
	}

	return true
}
