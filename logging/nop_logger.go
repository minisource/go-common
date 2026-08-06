package logging

// NopLogger is a Logger implementation that discards all output.
// It is used as the default logger when none is configured, so optional
// logging never panics and never writes log files.
type NopLogger struct{}

// Init implements Logger.
func (NopLogger) Init() {}

// Debug implements Logger.
func (NopLogger) Debug(Category, SubCategory, string, map[ExtraKey]interface{}) {}

// Debugf implements Logger.
func (NopLogger) Debugf(string, ...interface{}) {}

// Info implements Logger.
func (NopLogger) Info(Category, SubCategory, string, map[ExtraKey]interface{}) {}

// Infof implements Logger.
func (NopLogger) Infof(string, ...interface{}) {}

// Warn implements Logger.
func (NopLogger) Warn(Category, SubCategory, string, map[ExtraKey]interface{}) {}

// Warnf implements Logger.
func (NopLogger) Warnf(string, ...interface{}) {}

// Error implements Logger.
func (NopLogger) Error(Category, SubCategory, string, map[ExtraKey]interface{}) {}

// Errorf implements Logger.
func (NopLogger) Errorf(string, ...interface{}) {}

// Fatal implements Logger (does NOT exit — this is a no-op logger).
func (NopLogger) Fatal(Category, SubCategory, string, map[ExtraKey]interface{}) {}

// Fatalf implements Logger (does NOT exit — this is a no-op logger).
func (NopLogger) Fatalf(string, ...interface{}) {}
