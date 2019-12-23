package main

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
)

const (
	defaultServiceAccountName = "default"
)

func includeImagePullSecret(sa *corev1.ServiceAccount, secretName string) bool {
	for _, imagePullSecret := range sa.ImagePullSecrets {
		if imagePullSecret.Name == secretName {
			return true
		}
	}
	return false
}

type patch struct {
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

func getPatchString(sa *corev1.ServiceAccount, secretName string) ([]byte, error) {
	saPatch := patch{
		// copy the slice
		ImagePullSecrets: append([]corev1.LocalObjectReference(nil), sa.ImagePullSecrets...),
	}
	if !includeImagePullSecret(sa, secretName) {
		saPatch.ImagePullSecrets = append(saPatch.ImagePullSecrets, corev1.LocalObjectReference{Name: secretName})
	}
	return json.Marshal(saPatch)
}
