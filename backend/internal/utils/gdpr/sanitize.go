package gdpr

import "strings"

// PIIKeys contains the fields that should be scrubbed for GDPR compliance.
var PIIKeys = map[string]bool{
	"email":    true,
	"phone":    true,
	"address":  true,
	"name":     true,
	"password": true,
	"token":    true,
	"ssn":      true,
	"assignee": true,
	"reporter": true,
	"client":   true,
}

// SanitizePayload recursively removes sensitive fields from the input structure.
func SanitizePayload(payload interface{}) interface{} {
	switch v := payload.(type) {
	case map[string]interface{}:
		sanitized := make(map[string]interface{})
		for key, val := range v {
			if PIIKeys[strings.ToLower(key)] {
				continue // Omit sensitive keys entirely
			}
			sanitized[key] = SanitizePayload(val)
		}
		return sanitized
	case []interface{}:
		sanitized := make([]interface{}, len(v))
		for i, item := range v {
			sanitized[i] = SanitizePayload(item)
		}
		return sanitized
	default:
		return v
	}
}
