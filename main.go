package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	// Config
	configForce                bool          = true
	configDebug                bool          = false
	configManagedOnly          bool          = false
	configRunOnce              bool          = false
	configAllServiceAccount    bool          = false
	configDockerconfigjson     string        = ""
	configDockerConfigJSONPath string        = ""
	configSecretName           string        = "image-pull-secret" // default to image-pull-secret
	configExcludedNamespaces   string        = ""
	configServiceAccounts      string        = defaultServiceAccountName
	configLoopDuration         time.Duration = 10 * time.Second

	dockerConfigJSON string
)

const (
	annotationImagepullsecretPatcherExclude = "k8s.titansoft.com/imagepullsecret-patcher-exclude"
)

type k8sClient struct {
	clientset kubernetes.Interface
}

func main() {
	// parse flags
	flag.BoolVar(&configForce, "force", LookUpEnvOrBool("CONFIG_FORCE", configForce), "force to overwrite secrets when not match")
	flag.BoolVar(&configDebug, "debug", LookUpEnvOrBool("CONFIG_DEBUG", configDebug), "show DEBUG logs")
	flag.BoolVar(&configManagedOnly, "managedonly", LookUpEnvOrBool("CONFIG_MANAGEDONLY", configManagedOnly), "only modify secrets which are annotated as managed by imagepullsecret")
	flag.BoolVar(&configRunOnce, "runonce", LookUpEnvOrBool("CONFIG_RUNONCE", configRunOnce), "run a single update and exit instead of looping")
	flag.BoolVar(&configAllServiceAccount, "allserviceaccount", LookUpEnvOrBool("CONFIG_ALLSERVICEACCOUNT", configAllServiceAccount), "if false, patch just default service account; if true, list and patch all service accounts")
	flag.StringVar(&configDockerconfigjson, "dockerconfigjson", LookupEnvOrString("CONFIG_DOCKERCONFIGJSON", configDockerconfigjson), "json credential for authenicating container registry, exclusive with `dockerconfigjsonpath`")
	flag.StringVar(&configDockerConfigJSONPath, "dockerconfigjsonpath", LookupEnvOrString("CONFIG_DOCKERCONFIGJSONPATH", configDockerConfigJSONPath), "path to json file containing credentials for the registry to be distributed, exclusive with `dockerconfigjson`")
	flag.StringVar(&configSecretName, "secretname", LookupEnvOrString("CONFIG_SECRETNAME", configSecretName), "set name of managed secrets")
	flag.StringVar(&configExcludedNamespaces, "excluded-namespaces", LookupEnvOrString("CONFIG_EXCLUDED_NAMESPACES", configExcludedNamespaces), "comma-separated namespaces excluded from processing")
	flag.StringVar(&configServiceAccounts, "serviceaccounts", LookupEnvOrString("CONFIG_SERVICEACCOUNTS", configServiceAccounts), "comma-separated list of serviceaccounts to patch")
	flag.DurationVar(&configLoopDuration, "loop-duration", LookupEnvOrDuration("CONFIG_LOOP_DURATION", configLoopDuration), "String defining the loop duration")
	flag.Parse()

	// setup logrus
	if configDebug {
		log.SetLevel(log.DebugLevel)
	}
	log.Info("Application started")

	// Validate input, as both of these being configured would have undefined behavior.
	if configDockerconfigjson != "" && configDockerConfigJSONPath != "" {
		log.Panic(fmt.Errorf("Cannot specify both `configdockerjson` and `configdockerjsonpath`"))
	}

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
		if configRunOnce {
			log.Info("Exiting after single loop per `CONFIG_RUNONCE`")
			os.Exit(0)
		}
		time.Sleep(configLoopDuration)
	}
}

func loop(k8s *k8sClient) {
	var err error

	// Populate secret value to set
	dockerConfigJSON, err = getDockerConfigJSON()
	if err != nil {
		log.Panic(err)
	}

	// get all namespaces
	namespaces, err := k8s.clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		log.Panic(err)
	}
	log.Debugf("Got %d namespaces", len(namespaces.Items))

	for _, ns := range namespaces.Items {
		namespace := ns.Name
		if namespaceIsExcluded(ns) {
			log.Infof("[%s] Namespace skipped", namespace)
			continue
		}
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

func namespaceIsExcluded(ns corev1.Namespace) bool {
	v, ok := ns.Annotations[annotationImagepullsecretPatcherExclude]
	if ok && v == "true" {
		return true
	}
	for _, ex := range strings.Split(configExcludedNamespaces, ",") {
		if ex == ns.Name {
			return true
		}
	}
	return false
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
		if configManagedOnly && isManagedSecret(secret) {
			return fmt.Errorf("[%s] Secret is present but unmanaged", namespace)
		}
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
	sas, err := k8s.clientset.CoreV1().ServiceAccounts(namespace).List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("[%s] Failed to list service accounts: %v", namespace, err)
	}
	for _, sa := range sas.Items {
		if !configAllServiceAccount && stringNotInList(sa.Name, configServiceAccounts) {
			log.Debugf("[%s] Skip service account [%s]", namespace, sa.Name)
			continue
		}
		if includeImagePullSecret(&sa, configSecretName) {
			log.Debugf("[%s] ImagePullSecrets found", namespace)
			continue
		}
		patch, err := getPatchString(&sa, configSecretName)
		if err != nil {
			return fmt.Errorf("[%s] Failed to get patch string: %v", namespace, err)
		}
		_, err = k8s.clientset.CoreV1().ServiceAccounts(namespace).Patch(sa.Name, types.StrategicMergePatchType, patch)
		if err != nil {
			return fmt.Errorf("[%s] Failed to patch imagePullSecrets to service account [%s]: %v", namespace, sa.Name, err)
		}
		log.Infof("[%s] Patched imagePullSecrets to service account [%s]", namespace, sa.Name)
	}
	return nil
}

func stringNotInList(a string, list string) bool {
	for _, b := range strings.Split(list, ",") {
		if b == a {
			return false
		}
	}
	return true
}
