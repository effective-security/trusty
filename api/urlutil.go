package api

import (
	"net/url"
)

// GetQueryString returns Query parameter
func GetQueryString(u *url.URL, name string) string {
	vals, ok := u.Query()[name]
	if !ok || len(vals) == 0 {
		return ""
	}
	return vals[0]
}
