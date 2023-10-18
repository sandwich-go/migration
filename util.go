package migration

import (
	"strings"
)

func HidePassword(source interface{}, pass string) interface{} {
	if data, ok := source.(string); ok {
		return hide(data, pass, false)
	} else if dataSlice, ok := source.([]string); ok {
		var newDataSlice []string
		for _, str := range dataSlice {
			newDataSlice = append(newDataSlice, hide(str, pass, true))
		}
		return newDataSlice
	}
	return source
}

func hide(data, pass string, strict bool) string {
	if (strict && data == pass) || (!strict && strings.Contains(data, pass)) {
		return strings.ReplaceAll(data, pass, "******")
	}
	return data
}
