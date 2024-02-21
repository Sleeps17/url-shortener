package random

import (
	"math/rand"
)

const defaultAliasLength = 16

func Alias(length ...int) string {

	if len(length) > 1 {
		return ""
	}

	l := 0
	if len(length) == 1 {
		l = length[0]
	} else {
		l = defaultAliasLength
	}

	if l < 0 {
		return ""
	}

	symbols := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	alias := make([]byte, 0, l)

	for i := 0; i < l; i++ {
		alias = append(alias, symbols[rand.Intn(len(symbols))])
	}

	return string(alias)
}
