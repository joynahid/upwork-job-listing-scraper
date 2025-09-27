package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func clientIP(c *gin.Context) string {
	if ip := c.ClientIP(); ip != "" {
		return ip
	}
	return "unknown"
}

func mustEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable is required", key)
	}
	return value
}

func maskAPIKey(value string) string {
	if value == "" {
		return "(empty)"
	}
	if len(value) <= 4 {
		return strings.Repeat("*", len(value))
	}
	return fmt.Sprintf("%s***%s", value[:2], value[len(value)-2:])
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func ptrFloat(value float64) *float64 { return &value }

type serviceAccountPayload struct {
	ProjectID string `json:"project_id"`
}

func loadProjectID(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("unable to read service account file: %w", err)
	}

	var payload serviceAccountPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", fmt.Errorf("unable to parse service account file: %w", err)
	}

	if payload.ProjectID == "" {
		return "", fmt.Errorf("project_id not found in service account file")
	}

	return payload.ProjectID, nil
}

func getMap(root map[string]interface{}, keys ...string) map[string]interface{} {
	current := root
	for _, key := range keys {
		if current == nil {
			return nil
		}
		value, ok := current[key]
		if !ok {
			return nil
		}
		next, ok := value.(map[string]interface{})
		if !ok {
			return nil
		}
		current = next
	}
	return current
}

func firstNonNilMap(maps ...map[string]interface{}) map[string]interface{} {
	for _, m := range maps {
		if len(m) > 0 {
			return m
		}
	}
	return nil
}

func isValidJobMap(m map[string]interface{}) bool {
	if m == nil {
		return false
	}
	if _, ok := m["title"]; ok {
		return true
	}
	if _, ok := m["uid"]; ok {
		return true
	}
	if _, ok := m["description"]; ok {
		return true
	}
	return false
}

func getString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if value, ok := m[key]; ok {
		if s, ok := value.(string); ok {
			return s
		}
	}
	return ""
}

func firstString(m map[string]interface{}, paths ...[]string) (string, bool) {
	for _, path := range paths {
		if value, ok := dig(m, path...); ok {
			if str, ok := value.(string); ok && strings.TrimSpace(str) != "" {
				return str, true
			}
		}
	}
	return "", false
}

func dig(root interface{}, keys ...string) (interface{}, bool) {
	current := root
	for _, key := range keys {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		value, ok := m[key]
		if !ok {
			return nil, false
		}
		current = value
	}
	return current, true
}

func getIntPointer(m map[string]interface{}, key string) *int {
	if m == nil {
		return nil
	}
	if value, ok := m[key]; ok {
		if intVal, ok := toInt(value); ok {
			return &intVal
		}
	}
	return nil
}

func extractInt(m map[string]interface{}, key string) (int, bool) {
	if value, ok := dig(m, key); ok {
		if intVal, ok := toInt(value); ok {
			return intVal, true
		}
	}
	return 0, false
}

func extractFloat(m map[string]interface{}, key string) (float64, bool) {
	if value, ok := dig(m, key); ok {
		return toFloat64(value)
	}
	return 0, false
}

func extractBool(m map[string]interface{}, key string) (bool, bool) {
	if value, ok := dig(m, key); ok {
		switch v := value.(type) {
		case bool:
			return v, true
		case string:
			parsed, err := strconv.ParseBool(v)
			if err == nil {
				return parsed, true
			}
		}
	}
	return false, false
}

func extractStringSlice(root map[string]interface{}, keys ...string) ([]string, bool) {
	value, ok := dig(root, keys...)
	if !ok {
		return nil, false
	}
	arr, ok := value.([]interface{})
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if str, ok := item.(string); ok && str != "" {
			out = append(out, str)
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

func extractMapSlice(root map[string]interface{}, keys ...string) []map[string]interface{} {
	value, ok := dig(root, keys...)
	if !ok {
		return nil
	}
	arr, ok := value.([]interface{})
	if !ok {
		return nil
	}
	result := make([]map[string]interface{}, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]interface{}); ok && len(m) > 0 {
			result = append(result, m)
		}
	}
	return result
}

func firstTime(root map[string]interface{}, paths ...[]string) *time.Time {
	for _, path := range paths {
		if value, ok := dig(root, path...); ok {
			switch v := value.(type) {
			case time.Time:
				t := v.UTC()
				return &t
			case *time.Time:
				if v != nil {
					t := v.UTC()
					return &t
				}
			case string:
				if ts, err := parseFlexibleTime(v); err == nil {
					return &ts
				}
			}
		}
	}
	return nil
}

func toInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case float32:
		return int(v), true
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return int(i), true
		}
		if f, err := v.Float64(); err == nil {
			return int(f), true
		}
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, true
		}
	}
	return 0, false
}

func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return f, true
		}
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func timeOrZero(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.UTC()
}

func parseTimeParam(value string) (time.Time, error) {
	if ts, err := parseFlexibleTime(value); err == nil {
		return ts, nil
	}
	return time.Time{}, fmt.Errorf("invalid time format")
}

func parseFlexibleTime(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05.999999Z07:00",
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, value); err == nil {
			return ts.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized time format: %s", value)
}

func firstQuery(values url.Values, key string) string {
	if values == nil {
		return ""
	}
	if vs, ok := values[key]; ok && len(vs) > 0 {
		return vs[0]
	}
	return ""
}
