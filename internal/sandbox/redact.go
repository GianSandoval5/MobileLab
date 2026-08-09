package sandbox

import "strings"

const redacted = "[REDACTED]"

var sensitiveNames = map[string]struct{}{
	"authorization": {}, "cookie": {}, "set-cookie": {}, "password": {},
	"passwd": {}, "token": {}, "access_token": {}, "refresh_token": {},
	"api_key": {}, "apikey": {}, "x_api_key": {}, "secret": {},
}

func RedactHeaders(headers map[string][]string) map[string][]string {
	result := make(map[string][]string, len(headers))
	for key, values := range headers {
		if isSensitive(key) {
			result[key] = []string{redacted}
			continue
		}
		result[key] = append([]string(nil), values...)
	}
	return result
}

func RedactValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, nested := range typed {
			if isSensitive(key) {
				result[key] = redacted
			} else {
				result[key] = RedactValue(nested)
			}
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for index, nested := range typed {
			result[index] = RedactValue(nested)
		}
		return result
	default:
		return value
	}
}

func isSensitive(name string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), "-", "_"))
	_, sensitive := sensitiveNames[normalized]
	return sensitive || strings.Contains(normalized, "password") || strings.HasSuffix(normalized, "_token")
}
