package main

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var testCasesProcessSecret = []testCase{
	{
		name: "no secret",
		prepSteps: []step{
			assertNoSecret,
		},
		testSteps: []step{
			processSecretDefault,
			assertSecretIsValid,
		},
	},
	{
		name: "has valid secret",
		prepSteps: []step{
			helperCreateValidSecret,
			assertSecretIsValid,
		},
		testSteps: []step{
			processSecretDefault,
			assertSecretIsValid,
		},
	},
	{
		name: "has invalid secret - force on",
		prepSteps: []step{
			helperForceOn,
			helperCreateOpaqueSecret,
			assertSecretIsInvalid,
		},
		testSteps: []step{
			processSecretDefault,
			assertSecretIsValid,
		},
	},
	{
		name: "has invalid secret - force off",
		prepSteps: []step{
			helperForceOff,
			helperCreateOpaqueSecret,
			assertSecretIsInvalid,
		},
		testSteps: []step{
			assertHasError(processSecretDefault),
			assertSecretIsInvalid,
		},
	},
}

var testCasesProcessServiceAccount = []testCase{
	{
		name: "no image pull secret",
		prepSteps: []step{
			helperCreateServiceAccountWithoutImagePullSecret(defaultServiceAccountName),
			assertHasError(assertHasImagePullSecret(configSecretName, defaultServiceAccountName)),
		},
		testSteps: []step{
			processServiceAccountDefault,
			assertHasImagePullSecret(configSecretName, defaultServiceAccountName),
		},
	},
	{
		name: "has same image pull secret",
		prepSteps: []step{
			helperCreateServiceAccountWithImagePullSecret(configSecretName, defaultServiceAccountName),
			assertHasImagePullSecret(configSecretName, defaultServiceAccountName),
		},
		testSteps: []step{
			processServiceAccountDefault,
			assertHasImagePullSecret(configSecretName, defaultServiceAccountName),
		},
	},
	{
		name: "has different image pull secret",
		prepSteps: []step{
			helperCreateServiceAccountWithImagePullSecret("other-secret", defaultServiceAccountName),
			assertHasImagePullSecret("other-secret", defaultServiceAccountName),
			assertHasError(assertHasImagePullSecret(configSecretName, defaultServiceAccountName)),
		},
		testSteps: []step{
			processServiceAccountDefault,
			assertHasImagePullSecret("other-secret", defaultServiceAccountName),
			assertHasImagePullSecret(configSecretName, defaultServiceAccountName),
		},
	},
	{
		name: "non-default service account - skip when allServiceAccount off",
		prepSteps: []step{
			helperAllServiceAccountOff,
			helperCreateServiceAccountWithoutImagePullSecret("other-service-account"),
			assertHasError(assertHasImagePullSecret(configSecretName, "other-service-account")),
		},
		testSteps: []step{
			processServiceAccountDefault,
			assertHasError(assertHasImagePullSecret(configSecretName, "other-service-account")),
		},
	},
	{
		name: "non-default service account - patch when allServiceAccount on",
		prepSteps: []step{
			helperAllServiceAccountOn,
			helperCreateServiceAccountWithoutImagePullSecret("other-service-account"),
			assertHasError(assertHasImagePullSecret(configSecretName, "other-service-account")),
		},
		testSteps: []step{
			processServiceAccountDefault,
			assertHasImagePullSecret(configSecretName, "other-service-account"),
		},
	},
}

func TestProcessSecret(t *testing.T) {
	for _, tc := range testCasesProcessSecret {
		runTestCase(t, "ProcessSecret", tc)
	}
}

func TestProcessServiceAccount(t *testing.T) {
	for _, tc := range testCasesProcessServiceAccount {
		runTestCase(t, "ProcessServiceAccount", tc)
	}
}

type step func(*k8sClient) error

type testCase struct {
	name      string // name of the test
	prepSteps []step // preparation steps
	testSteps []step // test steps
}

func runTestCase(t *testing.T, testName string, tc testCase) {
	// disable logrus
	logrus.SetOutput(ioutil.Discard)

	// create fake client
	k8s := &k8sClient{
		clientset: fake.NewSimpleClientset(),
	}

	// run preparation steps
	for _, step := range tc.prepSteps {
		if err := step(k8s); err != nil {
			t.Errorf("%s(%s) failed during preparation: %v", testName, tc.name, err)
			return
		}
	}

	// run through test steps
	for _, step := range tc.testSteps {
		if err := step(k8s); err != nil {
			t.Errorf("%s(%s) failed during test: %v", testName, tc.name, err)
			return
		}
	}
}

func processSecretDefault(k8s *k8sClient) error {
	return processSecret(k8s, v1.NamespaceDefault)
}

func processServiceAccountDefault(k8s *k8sClient) error {
	return processServiceAccount(k8s, v1.NamespaceDefault)
}

func TestNamespaceIsExcluded(t *testing.T) {
	for _, tc := range []struct {
		name      string
		config    string
		namespace corev1.Namespace
		expected  bool
	}{
		{
			name:   "empty config",
			config: "",
			namespace: corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "kube-system",
				},
			},
			expected: false,
		},
		{
			name:   "appear in config",
			config: "kube-system,other-namespace",
			namespace: corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "kube-system",
				},
			},
			expected: true,
		},
		{
			name:   "not appear in config",
			config: "default,other-namespace",
			namespace: corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "kube-system",
				},
			},
			expected: false,
		},
		{
			name:   "namespace has annotation true",
			config: "",
			namespace: corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "kube-system",
					Annotations: map[string]string{
						"k8s.titansoft.com/imagepullsecret-patcher-exclude": "true",
					},
				},
			},
			expected: true,
		},
	} {
		configExcludedNamespaces = tc.config
		if actual := namespaceIsExcluded(tc.namespace); actual != tc.expected {
			t.Errorf("TestNamespaceIsExcluded(%s) failed: expected %v, got %v", tc.name, tc.expected, actual)
		}
	}
}

// a set of helper functions
func helperCreateValidSecret(k8s *k8sClient) error {
	_, err := k8s.clientset.CoreV1().Secrets(v1.NamespaceDefault).Create(dockerconfigSecret(v1.NamespaceDefault))
	return err
}

func helperCreateOpaqueSecret(k8s *k8sClient) error {
	_, err := k8s.clientset.CoreV1().Secrets(v1.NamespaceDefault).Create(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configSecretName,
			Namespace: v1.NamespaceDefault,
		},
		Type: corev1.SecretTypeOpaque,
	})
	return err
}

func helperCreateServiceAccountWithoutImagePullSecret(serviceAccountName string) step {
	return func(k8s *k8sClient) error {
		_, err := k8s.clientset.CoreV1().ServiceAccounts(v1.NamespaceDefault).Create(&v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceAccountName,
				Namespace: v1.NamespaceDefault,
			},
		})
		return err
	}
}

func helperCreateServiceAccountWithImagePullSecret(secretName, serviceAccountName string) step {
	return func(k8s *k8sClient) error {
		_, err := k8s.clientset.CoreV1().ServiceAccounts(v1.NamespaceDefault).Create(&v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceAccountName,
				Namespace: v1.NamespaceDefault,
			},
			ImagePullSecrets: []v1.LocalObjectReference{
				{
					Name: secretName,
				},
			},
		})
		return err
	}
}

func helperForceOn(_ *k8sClient) error {
	configForce = true
	return nil
}

func helperForceOff(_ *k8sClient) error {
	configForce = false
	return nil
}

func helperAllServiceAccountOn(_ *k8sClient) error {
	configAllServiceAccount = true
	return nil
}

func helperAllServiceAccountOff(_ *k8sClient) error {
	configAllServiceAccount = false
	return nil
}

// a set of assertion functions
func assertNoSecret(k8s *k8sClient) error {
	_, err := k8s.clientset.CoreV1().Secrets(v1.NamespaceDefault).Get(configSecretName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return nil
	}
	if err == nil {
		return fmt.Errorf("assert no secret but found")
	}
	return err
}

func assertSecretIsValid(k8s *k8sClient) error {
	secret, err := k8s.clientset.CoreV1().Secrets(v1.NamespaceDefault).Get(configSecretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("assert secret valid but no found")
	}
	if result := verifySecret(secret); result != secretOk {
		return fmt.Errorf("assert secret valid but invalid: %v", result)
	}
	return nil
}

func assertSecretIsInvalid(k8s *k8sClient) error {
	secret, err := k8s.clientset.CoreV1().Secrets(v1.NamespaceDefault).Get(configSecretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("assert secret invalid but no found")
	}
	if result := verifySecret(secret); result == secretOk {
		return fmt.Errorf("assert secret invalid but valid")
	}
	return nil
}

func assertHasError(fn step) step {
	return func(k8s *k8sClient) error {
		if err := fn(k8s); err == nil {
			return fmt.Errorf("assert has error but not")
		}
		return nil
	}
}

func assertHasImagePullSecret(secretName, serviceAccountName string) step {
	return func(k8s *k8sClient) error {
		sa, err := k8s.clientset.CoreV1().ServiceAccounts(v1.NamespaceDefault).Get(serviceAccountName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if includeImagePullSecret(sa, secretName) {
			return nil
		}
		return fmt.Errorf("assert has image pull secret [%s] but not found", secretName)
	}
}
