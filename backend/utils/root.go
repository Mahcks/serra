package utils

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

// B2S converts byte slice to a string without memory allocation.
func B2S(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// S2B converts a string to a byte slice without memory allocation.
func S2B(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// Ternary returns whenTrue if condition is true, otherwise whenFalse
func Ternary[T any](condition bool, whenTrue T, whenFalse T) T {
	if condition {
		return whenTrue
	}
	return whenFalse
}

// IsEmptyValue uses reflection to determine if a value is empty.
func IsEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// PrettyPrint prints a struct or map as pretty-formatted JSON
func PrettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to pretty-print JSON:", err)
		return
	}
	fmt.Println(string(b))
}

// GeneratePassword generates a random password of given length
func GeneratePassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	password := make([]byte, length)
	for i := range password {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password[i] = charset[num.Int64()]
	}
	return string(password), nil
}

// EmptyResult returns an empty slice and logs a warning
func EmptyResult[T any](logMsg string) ([]T, error) {
	slog.Warn(logMsg)
	return []T{}, nil
}

// Nullable wrappers for sql.Null* types with fluent API
type NullableString struct{ sql.NullString }
type NullableInt64 struct{ sql.NullInt64 }
type NullableFloat64 struct{ sql.NullFloat64 }
type NullableBool struct{ sql.NullBool }
type NullableTime struct{ sql.NullTime }

// Or returns the value if valid, otherwise returns defaultValue
func (ns NullableString) Or(defaultValue string) string {
	if ns.Valid {
		return ns.String
	}
	return defaultValue
}

func (ni NullableInt64) Or(defaultValue int64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return defaultValue
}

func (nf NullableFloat64) Or(defaultValue float64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return defaultValue
}

func (nb NullableBool) Or(defaultValue bool) bool {
	if nb.Valid {
		return nb.Bool
	}
	return defaultValue
}

func (nt NullableTime) Or(defaultValue time.Time) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return defaultValue
}

// ToPointer returns a pointer to the value if valid, nil otherwise
func (ns NullableString) ToPointer() *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func (ni NullableInt64) ToPointer() *int64 {
	if ni.Valid {
		return &ni.Int64
	}
	return nil
}

func (nf NullableFloat64) ToPointer() *float64 {
	if nf.Valid {
		return &nf.Float64
	}
	return nil
}

func (nb NullableBool) ToPointer() *bool {
	if nb.Valid {
		return &nb.Bool
	}
	return nil
}

func (nt NullableTime) ToPointer() *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// IsValid returns true if the nullable value is valid
func (ns NullableString) IsValid() bool  { return ns.Valid }
func (ni NullableInt64) IsValid() bool   { return ni.Valid }
func (nf NullableFloat64) IsValid() bool { return nf.Valid }
func (nb NullableBool) IsValid() bool    { return nb.Valid }
func (nt NullableTime) IsValid() bool    { return nt.Valid }

// Helper functions for creating sql.Null* types
func NewNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func NewNullInt64(i int64, valid bool) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: valid}
}

func NewNullFloat64(f float64, valid bool) sql.NullFloat64 {
	return sql.NullFloat64{Float64: f, Valid: valid}
}

func NewNullBool(b bool) sql.NullBool {
	return sql.NullBool{Bool: b, Valid: true}
}

func NewNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: !t.IsZero()}
}

// Simple pointer helpers
func PtrString(val string) *string   { return &val }
func PtrInt64(val int64) *int64      { return &val }
func PtrInt(val int) *int            { return &val }
func PtrFloat64(val float64) *float64 { return &val }
func PtrBool(val bool) *bool         { return &val }

func DerefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func DerefInt64(i *int64) int64 {
	if i != nil {
		return *i
	}
	return 0
}

func DerefInt(i *int) int {
	if i != nil {
		return *i
	}
	return 0
}

func DerefFloat64(f *float64) float64 {
	if f != nil {
		return *f
	}
	return 0
}

func DerefBool(b *bool) bool {
	if b != nil {
		return *b
	}
	return false
}

// URL building helper
func BuildURL(baseURL, path string, params map[string]string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL + path
	}
	
	u.Path = strings.TrimSuffix(u.Path, "/") + "/" + strings.TrimPrefix(path, "/")
	
	if len(params) > 0 {
		query := u.Query()
		for key, value := range params {
			if value != "" {
				query.Set(key, value)
			}
		}
		u.RawQuery = query.Encode()
	}
	
	return u.String()
}

// String utilities
func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func MatchesAnyPattern(s string, patterns []string, ignoreCase bool) bool {
	target := s
	if ignoreCase {
		target = strings.ToLower(s)
	}
	
	for _, pattern := range patterns {
		checkPattern := pattern
		if ignoreCase {
			checkPattern = strings.ToLower(pattern)
		}
		if strings.Contains(target, checkPattern) {
			return true
		}
	}
	return false
}

// Duration formatting
func FormatDuration(seconds int64) string {
	if seconds <= 0 {
		return "∞"
	}
	
	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

// Size formatting
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// Parse size from string (e.g., "1.5 GB" -> bytes)
func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
	parts := strings.Fields(sizeStr)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}
	
	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}
	
	unit := strings.ToUpper(parts[1])
	multipliers := map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}
	
	multiplier, ok := multipliers[unit]
	if !ok {
		return 0, fmt.Errorf("unknown size unit: %s", unit)
	}
	
	return int64(value * float64(multiplier)), nil
}

// HTTP status helpers
func IsHTTPSuccess(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

func IsHTTPError(statusCode int) bool {
	return statusCode >= 400
}

// Safe string to int conversion
func SafeAtoi(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return 0
}

// Safe string to int64 conversion
func SafeAtoi64(s string) int64 {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	return 0
}

// Safe string to float64 conversion
func SafeAtof(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return 0
}
