package migration

import (
	"strings"
)

func HidePassword(source interface{}, pass string) interface{} {
	if data, ok := source.(string); ok {
		return hide(data, pass)
	} else if dataSlice, ok := source.([]string); ok {
		var newDataSlice []string
		for _, str := range dataSlice {
			newDataSlice = append(newDataSlice, hide(str, pass))
		}
		return newDataSlice
	}
	return source
}

func hide(data, pass string) string {
	if !strings.Contains(data, pass) {
		return data
	}
	return strings.ReplaceAll(data, pass, "******")
}
