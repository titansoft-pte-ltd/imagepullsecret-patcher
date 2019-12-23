package main

import (
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func dockerconfigSecret(namespace string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      configSecretName,
			Namespace: namespace,
			Annotations: map[string]string{
				"app.kubernetes.io/managed-by": "gcr-cred-patcher",
			},
		},
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: []byte(configDockerconfigjson),
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}
}

type verifySecretResult string

const (
	// error code for verifySecret
	secretOk           verifySecretResult = "SecretOk"
	secretWrongType    verifySecretResult = "SecretWrongType"
	secretNoKey        verifySecretResult = "SecretNoKey"
	secretDataNotMatch verifySecretResult = "SecretDataNotMatch"
)

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
