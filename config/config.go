package config

// HTTPConfig holds the URL paths for the HTTP monitoring endpoints.
//
// Blocked:    Endpoint to get the list of long locks (e.g. "/blocked").
// Categories: Endpoint to get statistics grouped by categories (e.g. "/categories").
// Stack:      Endpoint prefix to get the stack trace by its ID (e.g. "/stacks/").
//
//	Note the trailing slash, since the stack ID is appended to this path.
type HTTPConfig struct {
	Blocked    string
	Categories string
	Stack      string
}

// MonitorConfig contains configuration settings for the lock monitor and its HTTP server.
//
// KeepRecords:     Whether to keep detailed lock records in memory after unlocking.
// DefaultCategory: The default category name assigned to locks if none is set.
// HTTPServerAddr:  Address and port where the HTTP server will listen (e.g. ":8080").
// HTTPEndpoints:   Struct holding URL paths for the HTTP monitoring endpoints.
type MonitorConfig struct {
	KeepRecords     bool
	DefaultCategory string
	HTTPServerAddr  string
	HTTPEndpoints   HTTPConfig
}

// Default returns a MonitorConfig instance with sane default values.
func Default() *MonitorConfig {
	return &MonitorConfig{
		KeepRecords:     false,
		DefaultCategory: "GLOBAL",
		HTTPServerAddr:  "0.0.0.0:8080",
		HTTPEndpoints: HTTPConfig{
			Blocked:    "/blocked",
			Categories: "/categories",
			Stack:      "/stacks/",
		},
	}
}
