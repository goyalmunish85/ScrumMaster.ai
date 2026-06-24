package gdpr

import (
	"reflect"
	"testing"
)

func TestSanitizePayload(t *testing.T) {
	input := map[string]interface{}{
		"id":          "123",
		"email":       "test@example.com",
		"assignee":    "Alice",
		"description": "Some issue",
		"nested": map[string]interface{}{
			"reporter": "Bob",
			"phone":    "123456789",
			"status":   "OPEN",
		},
		"list": []interface{}{
			map[string]interface{}{
				"name": "Charlie",
				"age":  30,
			},
			"normal_string",
		},
	}

	expected := map[string]interface{}{
		"id":          "123",
		"description": "Some issue",
		"nested": map[string]interface{}{
			"status": "OPEN",
		},
		"list": []interface{}{
			map[string]interface{}{
				"age": 30,
			},
			"normal_string",
		},
	}

	result := SanitizePayload(input)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("SanitizePayload failed.\nExpected: %v\nGot: %v", expected, result)
	}
}
