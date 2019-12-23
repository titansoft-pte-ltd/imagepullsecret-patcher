package main

import (
	"flag"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	configForce            bool   = false
	configDebug            bool   = false
	configDockerconfigjson string = ""
	configSecretName       string = "image-pull-secret" // default to image-pull-secret
)

type k8sClient struct {
	clientset kubernetes.Interface
}

func main() {
	// parse flags
	flag.BoolVar(&configForce, "force", LookUpEnvOrBool("CONFIG_FORCE", configForce), "force to overwrite secrets when not match")
	flag.BoolVar(&configDebug, "debug", LookUpEnvOrBool("CONFIG_DEBUG", configDebug), "show DEBUG logs")
	flag.StringVar(&configDockerconfigjson, "dockerconfigjson", LookupEnvOrString("CONFIG_DOCKERCONFIGJSON", configDockerconfigjson), "json credential for authenicating container registry")
	flag.StringVar(&configSecretName, "secretname", LookupEnvOrString("CONFIG_SECRETNAME", configSecretName), "set name of managed secrets")
	flag.Parse()

	// setup logrus
	if configDebug {
		log.SetLevel(log.DebugLevel)
	}
	log.Info("Application started")

	// create k8s clientset from in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panic(err)
	}
	k8s := &k8sClient{
		clientset: clientset,
	}

	for {
		log.Debug("Loop started")
		loop(k8s)
		time.Sleep(10 * time.Second)
	}
}

func loop(k8s *k8sClient) {
	// get all namespaces
	namespaces, err := k8s.clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		log.Panic(err)
	}
	log.Debugf("Got %d namespaces", len(namespaces.Items))

	for _, ns := range namespaces.Items {
		namespace := ns.Name
		log.Debugf("[%s] Start processing", namespace)
		// for each namespace, make sure the dockerconfig secret exists
		err = processSecret(k8s, namespace)
		if err != nil {
			// if has error in processing secret, should skip processing service account
			log.Error(err)
			continue
		}
		// get default service account, and patch image pull secret if not exist
		err = processServiceAccount(k8s, namespace)
		if err != nil {
			log.Error(err)
		}
	}
}

func processSecret(k8s *k8sClient, namespace string) error {
	secret, err := k8s.clientset.CoreV1().Secrets(namespace).Get(configSecretName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err := k8s.clientset.CoreV1().Secrets(namespace).Create(dockerconfigSecret(namespace))
		if err != nil {
			return fmt.Errorf("[%s] Failed to create secret: %v", namespace, err)
		}
		log.Infof("[%s] Created secret", namespace)
	} else if err != nil {
		return fmt.Errorf("[%s] Failed to GET secret: %v", namespace, err)
	} else {
		switch verifySecret(secret) {
		case secretOk:
			log.Debugf("[%s] Secret is valid", namespace)
		case secretWrongType, secretNoKey, secretDataNotMatch:
			if configForce {
				log.Warnf("[%s] Secret is not valid, overwritting now", namespace)
				err = k8s.clientset.CoreV1().Secrets(namespace).Delete(configSecretName, &metav1.DeleteOptions{})
				if err != nil {
					return fmt.Errorf("[%s] Failed to delete secret [%s]: %v", namespace, configSecretName, err)
				}
				log.Warnf("[%s] Deleted secret [%s]", namespace, configSecretName)
				_, err = k8s.clientset.CoreV1().Secrets(namespace).Create(dockerconfigSecret(namespace))
				if err != nil {
					return fmt.Errorf("[%s] Failed to create secret: %v", namespace, err)
				}
				log.Infof("[%s] Created secret", namespace)
			} else {
				return fmt.Errorf("[%s] Secret is not valid, set --force to true to overwrite", namespace)
			}
		}
	}
	return nil
}

func processServiceAccount(k8s *k8sClient, namespace string) error {
	sa, err := k8s.clientset.CoreV1().ServiceAccounts(namespace).Get(defaultServiceAccountName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("[%s] Failed to get service account [%s]: %v", namespace, defaultServiceAccountName, err)
	}
	if includeImagePullSecret(sa, configSecretName) {
		log.Debugf("[%s] ImagePullSecrets found", namespace)
		return nil
	}
	patch, err := getPatchString(sa, configSecretName)
	if err != nil {
		return fmt.Errorf("[%s] Failed to get patch string: %v", namespace, err)
	}
	_, err = k8s.clientset.CoreV1().ServiceAccounts(namespace).Patch(defaultServiceAccountName, types.StrategicMergePatchType, patch)
	if err != nil {
		return fmt.Errorf("[%s] Failed to patch imagePullSecrets to service account [%s]: %v", namespace, defaultServiceAccountName, err)
	}
	log.Infof("[%s] Patched imagePullSecrets to service account [%s]", namespace, defaultServiceAccountName)
	return nil
}
