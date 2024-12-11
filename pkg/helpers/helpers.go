package helpers

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Contains checks if item is in slice
func Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// MatchPattern checks if a string matches a given pattern
func MatchPattern(value, pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(value)
}

// UniqueItems checks if all items in the array are unique
func UniqueItems(arr []interface{}) bool {
	seen := make(map[interface{}]bool)
	for _, item := range arr {
		if seen[item] {
			return false
		}
		seen[item] = true
	}
	return true
}

// IsHostnameValid validates hostname format
func IsHostnameValid(value interface{}) bool {
	hostname, ok := value.(string)
	if !ok {
		return false
	}
	if len(hostname) > 253 || len(hostname) == 0 {
		return false
	}
	regex := regexp.MustCompile(`^([a-zA-Z0-9](-?[a-zA-Z0-9])*\.)*[a-zA-Z]{2,}$`)
	return regex.MatchString(hostname)
}

// IsInt32 checks if value is within int32 range
func IsInt32(value interface{}) bool {
	const minInt32, maxInt32 = -2147483648, 2147483647
	switch v := value.(type) {
	case string:
		if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
			return IsInt32(floatVal)
		}
	case float64:
		return v >= minInt32 && v <= maxInt32 && v == float64(int32(v))
	case int:
		return v >= minInt32 && v <= maxInt32
	}
	return false
}

// IsInt64 checks if value is within int64 range
func IsInt64(value interface{}) bool {
	const minInt64, maxInt64 = -9223372036854775808, 9223372036854775807
	switch v := value.(type) {
	case string:
		if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
			return IsInt64(floatVal)
		}
	case float64:
		return v >= minInt64 && v <= maxInt64 && v == float64(int64(v))
	case int:
		return v >= minInt64 && v <= maxInt64
	}
	return false
}

// IsFloat checks if value is a float
func IsFloat(value interface{}) bool {
	_, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 32)
	return err == nil
}

// IsDouble checks if value is a double
func IsDouble(value interface{}) bool {
	_, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64)
	return err == nil
}

// IsBoolean validates boolean values
func IsBoolean(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return true
	case string:
		lower := strings.ToLower(v)
		return lower == "true" || lower == "false"
	default:
		return false
	}
}

// IsISO8601 validates ISO8601 date format
func IsISO8601(value interface{}) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	_, err := time.Parse(time.RFC3339, str)
	return err == nil
}

// IsString validates string values
func IsString(value interface{}) bool {
	_, ok := value.(string)
	return ok
}

// IsUUID validates UUID format
func IsUUID(value interface{}) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	regex := regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
	return regex.MatchString(str)
}

// IsEmail validates email format
func IsEmail(value interface{}) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	regex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return regex.MatchString(str)
}

// IsURL validates URL format
func IsURL(value interface{}) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	regex := regexp.MustCompile(`^(https?|ftp)://[^\s/$.?#].[^\s]*$`)
	return regex.MatchString(str)
}

// IsArray checks if value is an array or slice
func IsArray(value interface{}) bool {
	switch value.(type) {
	case []interface{}, []string, []int, []float64, []bool:
		return true
	default:
		return false
	}
}

// IsIPv4 validates IPv4 address
func IsIPv4(value interface{}) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	ip := net.ParseIP(str)
	return ip != nil && strings.Count(str, ":") == 0
}

// IsIPv6 validates IPv6 address
func IsIPv6(value interface{}) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	ip := net.ParseIP(str)
	return ip != nil && strings.Count(str, ":") > 1
}

// IsByte validates if value is a byte array
func IsByte(value interface{}) bool {
	_, ok := value.([]byte)
	return ok
}

// ParseNumber converts a string to a float64
func ParseNumber(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}

// ParseDuration parses duration string with support for common formats
func ParseDuration(duration string) (time.Duration, error) {
	return time.ParseDuration(duration)
}

// HashKeyMD5Base64 generates a base64 encoded MD5 hash of a string
func HashKeyMD5Base64(key string) string {
	hash := md5.Sum([]byte(key))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// TypeRegexMap contains regex patterns for common types
var TypeRegexMap = map[string]string{
	"string":   "[^/?#]+",
	"integer":  "\\d+",
	"int32":    "\\d+",
	"int64":    "\\d+",
	"number":   "\\d+(\\.\\d+)?",
	"boolean":  "true|false",
	"ISO8601":  "\\d{4}-\\d{2}-\\d{2}(T\\d{2}:\\d{2}:\\d{2}Z)?",
	"binary":   ".*",
	"password": ".*",
	"email":    "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}",
	"uuid":     "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}",
	"uri":      "[a-zA-Z][a-zA-Z0-9+.-]*:[^\\s]*",
	"hostname": "[a-zA-Z0-9.-]+",
	"ipv4":     "((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)",
	"ipv6": `(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|` +
		`([0-9a-fA-F]{1,4}:){1,7}:|` +
		`([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|` +
		`([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|` +
		`([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|` +
		`([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|` +
		`([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|` +
		`[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|` +
		`:((:[0-9a-fA-F]{1,4}){1,7}|:)` +
		`)`,
}

// SanitizeString replaces special characters in a string
func SanitizeString(value string) string {
	// Replace slashes with underscores and remove other special characters
	sanitized := strings.ReplaceAll(value, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, ":", "")
	sanitized = strings.ReplaceAll(sanitized, "?", "")
	sanitized = strings.ReplaceAll(sanitized, "&", "")
	return sanitized
}
