package rest

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

func SimpleQuery(q url.Values) map[string]string {
	result := make(map[string]string)

	for k, v := range q {
		if len(v) != 0 {
			result[strings.ToLower(k)] = v[0]
		}
	}

	return result
}

func GetPageSize(queryString string) (int, error) {
	pageSize := 0

	if queryString != "" {
		s, err := strconv.Atoi(queryString)
		if err != nil {
			return pageSize, errors.New("page size must be int type")
		}

		pageSize = s
	}

	return pageSize, nil
}
