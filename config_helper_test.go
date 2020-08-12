package main

import (
	"os"
	"testing"
	"time"
)

var testCasesLookupEnvOrString = []struct {
	name       string
	envs       map[string]string
	defaultVal string
	lookupKey  string
	expected   string
}{
	{
		name: "hit",
		envs: map[string]string{
			"TEST": "test",
		},
		lookupKey:  "TEST",
		defaultVal: "default",
		expected:   "test",
	},
	{
		name: "miss",
		envs: map[string]string{
			"MISS": "miss",
		},
		lookupKey:  "TEST",
		defaultVal: "default",
		expected:   "default",
	},
}

func TestLookupEnvOrString(t *testing.T) {
	for _, testCase := range testCasesLookupEnvOrString {
		prepareEnvs(testCase.envs)
		actual := LookupEnvOrString(testCase.lookupKey, testCase.defaultVal)
		if actual != testCase.expected {
			t.Errorf("LookupEnvOrString(%s) gives %s, expects %s", testCase.name, actual, testCase.expected)
		}
	}
}

var testCasesLookupEnvOrInt = []struct {
	name       string
	envs       map[string]string
	defaultVal int
	lookupKey  string
	expected   int
}{
	{
		name: "hit",
		envs: map[string]string{
			"TEST": "666",
		},
		lookupKey:  "TEST",
		defaultVal: 23333,
		expected:   666,
	},
	{
		name: "miss",
		envs: map[string]string{
			"MISS": "miss",
		},
		lookupKey:  "TEST",
		defaultVal: 23333,
		expected:   23333,
	},
	{
		name: "nan",
		envs: map[string]string{
			"TEST": "test",
		},
		lookupKey:  "TEST",
		defaultVal: 23333,
		expected:   23333,
	},
}

func TestLookupEnvOrInt(t *testing.T) {
	for _, testCase := range testCasesLookupEnvOrInt {
		prepareEnvs(testCase.envs)
		actual := LookupEnvOrInt(testCase.lookupKey, testCase.defaultVal)
		if actual != testCase.expected {
			t.Errorf("LookupEnvOrInt(%s) gives %d, expects %d", testCase.name, actual, testCase.expected)
		}
	}
}

var testCasesLookupEnvOrBool = []struct {
	name       string
	envs       map[string]string
	defaultVal bool
	lookupKey  string
	expected   bool
}{
	{
		name: "hit",
		envs: map[string]string{
			"TEST": "false",
		},
		lookupKey:  "TEST",
		defaultVal: true,
		expected:   false,
	},
	{
		name: "miss",
		envs: map[string]string{
			"MISS": "true",
		},
		lookupKey:  "TEST",
		defaultVal: true,
		expected:   true,
	},
	{
		name: "not a bool",
		envs: map[string]string{
			"TEST": "test",
		},
		lookupKey:  "TEST",
		defaultVal: true,
		expected:   true,
	},
}

func TestLookupEnvOrBool(t *testing.T) {
	for _, testCase := range testCasesLookupEnvOrBool {
		prepareEnvs(testCase.envs)
		actual := LookUpEnvOrBool(testCase.lookupKey, testCase.defaultVal)
		if actual != testCase.expected {
			t.Errorf("LookupEnvOrBool(%s) gives %v, expects %v", testCase.name, actual, testCase.expected)
		}
	}
}

var testCasesLookupEnvOrDuration = []struct {
	name       string
	envs       map[string]string
	defaultVal time.Duration
	lookupKey  string
	expected   time.Duration
}{
	{
		name: "hit",
		envs: map[string]string{
			"TEST": "30s",
		},
		lookupKey:  "TEST",
		defaultVal: 10 * time.Second,
		expected:   30 * time.Second,
	},
	{
		name: "miss",
		envs: map[string]string{
			"MISS": "30s",
		},
		lookupKey:  "TEST",
		defaultVal: 10 * time.Second,
		expected:   10 * time.Second,
	},
	{
		name: "not a time duration",
		envs: map[string]string{
			"TEST": "no duration string",
		},
		lookupKey:  "TEST",
		defaultVal: 10 * time.Second,
		expected:   10 * time.Second,
	},
}

func TestLookupEnvOrDuration(t *testing.T) {
	for _, testCase := range testCasesLookupEnvOrDuration {
		prepareEnvs(testCase.envs)
		actual := LookupEnvOrDuration(testCase.lookupKey, testCase.defaultVal)
		if actual != testCase.expected {
			t.Errorf("LookupEnvOrDuration(%s) gives %v, expects %v", testCase.name, actual, testCase.expected)
		}
	}
}

func prepareEnvs(envs map[string]string) {
	os.Clearenv()
	for k, v := range envs {
		os.Setenv(k, v)
	}
}
