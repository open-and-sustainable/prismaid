// Package config provides functionality to load and manage configuration settings
// for projects utilizing various AI model providers. The configuration is designed
// to be loaded from TOML files, matching the defined structures.
//
// The package supports configuration for different AI providers, authentication methods,
// and model-specific settings. It handles validation of input parameters, offers
// methods to access specific provider configurations, and provides fallback options
// when specific settings are not defined. Users can override default settings
// through environment variables or by specifying alternative configuration files.
package config
