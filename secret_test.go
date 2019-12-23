package main

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

const (
	testDockerconfig = `{"auth":{"gcr.io":{"username":"_json_key","password":"{}"}}}`
)

var testCasesVerifySecret = []struct {
	name     string
	input    *corev1.Secret
	expected verifySecretResult
}{
	{
		name: "valid",
		input: &corev1.Secret{
			Type: corev1.SecretTypeDockerConfigJson,
			Data: map[string][]byte{
				corev1.DockerConfigJsonKey: []byte(testDockerconfig),
			},
		},
		expected: secretOk,
	},
	{
		name: "invalid secret type",
		input: &corev1.Secret{
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				corev1.DockerConfigJsonKey: []byte(testDockerconfig),
			},
		},
		expected: secretWrongType,
	},
	{
		name: "invalid secret key",
		input: &corev1.Secret{
			Type: corev1.SecretTypeDockerConfigJson,
			Data: map[string][]byte{
				"test": []byte(testDockerconfig),
			},
		},
		expected: secretNoKey,
	},
	{
		name: "invalid secret value",
		input: &corev1.Secret{
			Type: corev1.SecretTypeDockerConfigJson,
			Data: map[string][]byte{
				corev1.DockerConfigJsonKey: []byte(`{"auth":"invalid"}`),
			},
		},
		expected: secretDataNotMatch,
	},
}

func TestVerifySecret(t *testing.T) {
	configDockerconfigjson = testDockerconfig
	for _, testCase := range testCasesVerifySecret {
		actual := verifySecret(testCase.input)
		if actual != testCase.expected {
			t.Errorf("verifySecret(%s) gives %s, expects %s", testCase.name, actual, testCase.expected)
		}
	}
}

func TestDockerconfigSecretIsValid(t *testing.T) {
	result := verifySecret(dockerconfigSecret("default"))
	if result != secretOk {
		t.Errorf("dockerconfigSecret generates invalid secret: %s", result)
	}
}
