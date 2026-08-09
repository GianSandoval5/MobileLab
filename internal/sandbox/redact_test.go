package sandbox

import "testing"

func TestRedactHeadersAndNestedValues(t *testing.T) {
	headers := RedactHeaders(map[string][]string{
		"Authorization": {"Bearer secret"},
		"X-Request-ID":  {"safe"},
		"X-API-Key":     {"secret"},
	})
	if headers["Authorization"][0] != redacted || headers["X-API-Key"][0] != redacted {
		t.Fatalf("sensitive headers not redacted: %v", headers)
	}
	if headers["X-Request-ID"][0] != "safe" {
		t.Fatalf("safe header unexpectedly changed: %v", headers)
	}

	value := RedactValue(map[string]any{
		"profile":      map[string]any{"password": "secret", "name": "safe"},
		"access_token": "secret",
	}).(map[string]any)
	profile := value["profile"].(map[string]any)
	if value["access_token"] != redacted || profile["password"] != redacted || profile["name"] != "safe" {
		t.Fatalf("unexpected redacted value: %#v", value)
	}
}
