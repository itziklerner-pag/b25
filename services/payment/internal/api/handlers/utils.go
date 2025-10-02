package handlers

import (
	"strconv"
)

func parseIntQuery(s string, target *int) (int, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	*target = val
	return val, nil
}
