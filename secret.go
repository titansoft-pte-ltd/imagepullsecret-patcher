package main

import (
	"io/ioutil"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type verifySecretResult string

const (
	// annotation constants
	annotationManagedBy = "app.kubernetes.io/managed-by"
	annotationAppName   = "imagepullsecret-patcher"

	// result code for verifySecret
	secretOk           verifySecretResult = "SecretOk"
	secretWrongType    verifySecretResult = "SecretWrongType"
	secretNoKey        verifySecretResult = "SecretNoKey"
	secretDataNotMatch verifySecretResult = "SecretDataNotMatch"
)

// getDockerConfigJSON is a dynamic getter for our secret value. It lets us
// dynamically fetch the value from file or return the hard coded value,
// providing a consistent interface for access
func getDockerConfigJSON() (string, error) {
	if configDockerConfigJSONPath != "" {
		b, ok := ioutil.ReadFile(configDockerConfigJSONPath)
		return string(b), ok
	}
	return configDockerconfigjson, nil
}

func dockerconfigSecret(namespace string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      configSecretName,
			Namespace: namespace,
			Annotations: map[string]string{
				annotationManagedBy: annotationAppName,
			},
		},
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: []byte(dockerConfigJSON),
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}
}

func verifySecret(secret *corev1.Secret) verifySecretResult {
	if secret.Type != corev1.SecretTypeDockerConfigJson {
		return secretWrongType
	}
	b, ok := secret.Data[corev1.DockerConfigJsonKey]
	if !ok {
		return secretNoKey
	}
	if string(b) != dockerConfigJSON {
		return secretDataNotMatch
	}
	return secretOk
}

func isManagedSecret(secret *corev1.Secret) bool {
	if k, ok := secret.ObjectMeta.Annotations[annotationManagedBy]; ok {
		if k == annotationAppName {
			return true
		}
	}
	return false
}
