package main

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

var testCasesIncludeImagePullSecret = []struct {
	name       string
	sa         *corev1.ServiceAccount
	secretName string
	expected   bool
}{
	{
		name: "positive one secret",
		sa: &corev1.ServiceAccount{
			ImagePullSecrets: []corev1.LocalObjectReference{
				{Name: "secret-a"}}},
		secretName: "secret-a",
		expected:   true,
	},
	{
		name: "positive two secrets",
		sa: &corev1.ServiceAccount{
			ImagePullSecrets: []corev1.LocalObjectReference{
				{Name: "secret-a"},
				{Name: "secret-b"}}},
		secretName: "secret-a",
		expected:   true,
	},
	{
		name: "negative no secret",
		sa: &corev1.ServiceAccount{
			ImagePullSecrets: []corev1.LocalObjectReference{}},
		secretName: "secret-a",
		expected:   false,
	},
	{
		name: "negative one secret",
		sa: &corev1.ServiceAccount{
			ImagePullSecrets: []corev1.LocalObjectReference{
				{Name: "secret-b"}}},
		secretName: "secret-a",
		expected:   false,
	},
}

func TestIncludeImagePullSecret(t *testing.T) {
	for _, testCase := range testCasesIncludeImagePullSecret {
		actual := includeImagePullSecret(testCase.sa, testCase.secretName)
		if actual != testCase.expected {
			t.Errorf("includeImagePullSecret(%s) gives %v, expects %v", testCase.name, actual, testCase.expected)
		}
	}
}

var testCasesGetPatchString = []struct {
	name       string
	sa         *corev1.ServiceAccount
	secretName string
	expected   []byte
}{
	{
		name: "empty",
		sa: &corev1.ServiceAccount{
			ImagePullSecrets: []corev1.LocalObjectReference{}},
		secretName: "secret-a",
		expected:   []byte(`{"imagePullSecrets":[{"name":"secret-a"}]}`),
	},
	{
		name: "same",
		sa: &corev1.ServiceAccount{
			ImagePullSecrets: []corev1.LocalObjectReference{
				{Name: "secret-a"}}},
		secretName: "secret-a",
		expected:   []byte(`{"imagePullSecrets":[{"name":"secret-a"}]}`),
	},
	{
		name: "different",
		sa: &corev1.ServiceAccount{
			ImagePullSecrets: []corev1.LocalObjectReference{
				{Name: "secret-b"}}},
		secretName: "secret-a",
		expected:   []byte(`{"imagePullSecrets":[{"name":"secret-b"},{"name":"secret-a"}]}`),
	},
}

func TestGetPatchString(t *testing.T) {
	for _, testCase := range testCasesGetPatchString {
		actual, err := getPatchString(testCase.sa, testCase.secretName)
		if err != nil {
			t.Errorf("getPatchString(%s) has error %v", testCase.name, err)
		}
		if string(actual) != string(testCase.expected) {
			t.Errorf("getPatchString(%s) gives %s, expects %s", testCase.name, actual, testCase.expected)
		}
	}
}
