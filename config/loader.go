package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// Loader handles configuration loading from various sources
type Loader struct {
	envFiles  []string
	prefix    string
	loaded    bool
	loadOnce  sync.Once
	loadError error
}

// NewLoader creates a new config loader
func NewLoader() *Loader {
	return &Loader{
		envFiles: []string{".env"},
		prefix:   "",
	}
}

// WithEnvFiles sets the env files to load
func (l *Loader) WithEnvFiles(files ...string) *Loader {
	l.envFiles = files
	return l
}

// WithPrefix sets environment variable prefix
func (l *Loader) WithPrefix(prefix string) *Loader {
	l.prefix = prefix
	return l
}

// Load loads environment variables from files
func (l *Loader) Load() error {
	l.loadOnce.Do(func() {
		for _, file := range l.envFiles {
			if _, err := os.Stat(file); err == nil {
				if err := godotenv.Load(file); err != nil {
					l.loadError = fmt.Errorf("failed to load %s: %w", file, err)
					return
				}
			}
		}
		l.loaded = true
	})
	return l.loadError
}

// LoadInto loads configuration into a struct
func (l *Loader) LoadInto(cfg interface{}) error {
	if err := l.Load(); err != nil {
		return err
	}

	return unmarshalEnv(cfg, l.prefix)
}

// unmarshalEnv loads environment variables into a struct
func unmarshalEnv(cfg interface{}, prefix string) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("config must be a non-nil pointer to struct")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to struct")
	}

	return parseStruct(v, prefix)
}

func parseStruct(v reflect.Value, prefix string) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanSet() {
			continue
		}

		// Handle nested structs
		if field.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Time{}) {
			nestedPrefix := prefix
			if tag := fieldType.Tag.Get("env_prefix"); tag != "" {
				nestedPrefix = tag
			} else if prefix != "" {
				nestedPrefix = prefix + "_" + toSnakeCase(fieldType.Name)
			} else {
				nestedPrefix = toSnakeCase(fieldType.Name)
			}
			if err := parseStruct(field, nestedPrefix); err != nil {
				return err
			}
			continue
		}

		// Get env tag
		envKey := fieldType.Tag.Get("env")
		if envKey == "" {
			envKey = toSnakeCase(fieldType.Name)
		}

		if prefix != "" {
			envKey = prefix + "_" + envKey
		}
		envKey = strings.ToUpper(envKey)

		// Get value from environment
		envValue := os.Getenv(envKey)
		if envValue == "" {
			// Check for default tag
			if defaultVal := fieldType.Tag.Get("default"); defaultVal != "" {
				envValue = defaultVal
			} else {
				continue
			}
		}

		// Set the field value
		if err := setField(field, envValue); err != nil {
			return fmt.Errorf("failed to set field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

func setField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(value)
			if err != nil {
				// Try parsing as seconds
				secs, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return err
				}
				d = time.Duration(secs) * time.Second
			}
			field.Set(reflect.ValueOf(d))
		} else {
			i, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(i)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(u)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)

	case reflect.Slice:
		elemType := field.Type().Elem()
		if elemType.Kind() == reflect.String {
			parts := strings.Split(value, ",")
			slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))
			for i, part := range parts {
				slice.Index(i).SetString(strings.TrimSpace(part))
			}
			field.Set(slice)
		}

	default:
		return fmt.Errorf("unsupported type: %s", field.Kind())
	}

	return nil
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToUpper(result.String())
}

// ============================================
// Helper functions for manual config loading
// ============================================

// GetEnv gets an environment variable with default
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt gets an int environment variable with default
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

// GetEnvInt64 gets an int64 environment variable with default
func GetEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
	}
	return defaultValue
}

// GetEnvBool gets a bool environment variable with default
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

// GetEnvDuration gets a duration environment variable with default
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
		// Try parsing as seconds
		if secs, err := strconv.ParseInt(value, 10, 64); err == nil {
			return time.Duration(secs) * time.Second
		}
	}
	return defaultValue
}

// GetEnvSlice gets a comma-separated slice
func GetEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaultValue
}

// MustGetEnv gets an environment variable or panics
func MustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return value
}

// RequiredEnv checks if all required env vars are set
func RequiredEnv(keys ...string) error {
	missing := make([]string, 0)
	for _, key := range keys {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}
