package random

import (
	"fmt"
	"math/rand"
)

const defaultAliasLength = 16

func Alias(length ...int) (string, error) {
	const op = "random.Alias"

	if len(length) > 1 {
		return "", fmt.Errorf("%s: too many arguments", op)
	}

	l := 0
	if len(length) == 1 {
		l = length[0]
	} else {
		l = defaultAliasLength
	}

	symbols := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	alias := make([]byte, 0, l)

	for i := 0; i < l; i++ {
		alias = append(alias, symbols[rand.Intn(len(symbols))])
	}

	return string(alias), nil
}
