package main

import (
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
			corev1.DockerConfigJsonKey: []byte(configDockerconfigjson),
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
	if string(b) != configDockerconfigjson {
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
