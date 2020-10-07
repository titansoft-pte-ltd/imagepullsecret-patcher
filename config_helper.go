package main

import (
	"os"
	"strconv"
	"time"
)

// Reference:
// https://www.gmarik.info/blog/2019/12-factor-golang-flag-package/

// LookupEnvOrString lookup ENV string with given key,
// or returns default value if not exists
func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

// LookupEnvOrInt lookup ENV string with given key and convert to int,
// or returns default value if not exists or conversion failed
func LookupEnvOrInt(key string, defaultVal int) int {
	str, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	val, err := strconv.Atoi(str)
	if err != nil {
		return defaultVal
	}
	return val
}

// LookUpEnvOrBool lookup ENV string with given key and convert to bool,
// or returns default value if not exists or conversion failed
func LookUpEnvOrBool(key string, defaultVal bool) bool {
	str, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	val, err := strconv.ParseBool(str)
	if err != nil {
		return defaultVal
	}
	return val
}

// LookupEnvOrDuration lookup ENV string with given key and convert to time.Duration
// or returns default value if not exists
func LookupEnvOrDuration(key string, defaultVal time.Duration) time.Duration {
	str, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}

	val, err := time.ParseDuration(str)
	if err != nil {
		return defaultVal
	}

	return val
}
