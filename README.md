# imagepullsecret-patcher

[![Build Status](https://travis-ci.org/titansoft-pte-ltd/imagepullsecret-patcher.svg?branch=master)](https://travis-ci.org/titansoft-pte-ltd/imagepullsecret-patcher)
[![Go Report Card](https://goreportcard.com/badge/github.com/titansoft-pte-ltd/imagepullsecret-patcher)](https://goreportcard.com/report/github.com/titansoft-pte-ltd/imagepullsecret-patcher)
![Codecov](https://img.shields.io/codecov/c/github/titansoft-pte-ltd/imagepullsecret-patcher)
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/titansoft-pte-ltd/imagepullsecret-patcher)
![GitHub issues](https://img.shields.io/github/issues/titansoft-pte-ltd/imagepullsecret-patcher)

A simple Kubernetes [client-go](https://github.com/kubernetes/client-go) application that creates and patches imagePullSecrets to service accounts in all Kubernetes namespaces to allow cluster-wide authenticated access to private container registry.

![screenshot](doc/screenshot.png)

A blog post: https://medium.com/titansoft-engineering/kubernetes-cluster-wide-access-to-private-container-registry-with-imagepullsecret-patcher-b8b8fb79f7e5

## Installation and configuration

To install imagepullsecret-patcher, can refer to [deploy-example](deploy-example) as a quick-start.

Below is a table of available configurations:

| Config name          | ENV                         | Command flag          | Default value       | Description                                                                                                                      |
| -------------------- | --------------------------- | --------------------- | ------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| force                | CONFIG_FORCE                | -force                | true                | overwrite secrets when not match                                                                                                 |
| debug                | CONFIG_DEBUG                | -debug                | false               | show DEBUG logs                                                                                                                  |
| managedonly          | CONFIG_MANAGEDONLY          | -managedonly          | false               | only modify secrets which were created by imagepullsecret                                                                        |
| runonce              | CONFIG_RUNONCE              | -runonce              | false               | run the update loop once, allowing for cronjob scheduling if desired                                                             |
| serviceaccounts      | CONFIG_SERVICEACCOUNTS      | -serviceaccounts      | "default"           | comma-separated list of serviceaccounts to patch                                                                                 |
| all service account  | CONFIG_ALLSERVICEACCOUNT    | -allserviceaccount    | false               | if true, list and patch all service accounts and the `-servicesaccounts` argument is ignored                                     |
| dockerconfigjson     | CONFIG_DOCKERCONFIGJSON     | -dockerconfigjson     | ""                  | json credential for authenicating container registry                                                                             |
| dockerconfigjsonpath | CONFIG_DOCKERCONFIGJSONPATH | -dockerconfigjsonpath | ""                  | path for of mounted json credentials for dynamic secret management                                                               |
| secret name          | CONFIG_SECRETNAME           | -secretname           | "image-pull-secret" | name of managed secrets                                                                                                          |
| excluded namespaces  | CONFIG_EXCLUDED_NAMESPACES  | -excluded-namespaces  | ""                  | comma-separated namespaces excluded from processing                                                                              |
| loop duration        | CONFIG_LOOP_DURATION        | -loop-duration        | 10 seconds          | duration string which defines how often namespaces are checked, see https://golang.org/pkg/time/#ParseDuration for more examples |

And here are the annotations available:

| Annotation                                        | Object    | Description                                                                                                        |
| ------------------------------------------------- | --------- | ------------------------------------------------------------------------------------------------------------------ |
| k8s.titansoft.com/imagepullsecret-patcher-exclude | namespace | If a namespace is set this annotation with "true", it will be excluded from processing by imagepullsecret-patcher. |

## Providing credentials

You can provide the authentication credentials for imagepullsecret to populate across namespaces in a couple of ways.

You can provide a raw secret as an environment variable, or better yet, by mounting a volume into the container. Mounted secrets can be dynamically updated and are more secure. Please see the relevant docs for more information https://kubernetes.io/docs/concepts/configuration/secret/

## Why

To deploy private images to Kubernetes, we need to provide the credential to the private docker registries in either

- Pod definition (https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod)
- Default service account in a namespace (https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#add-imagepullsecrets-to-a-service-account)

With the second approach, a Kubernetes cluster admin configures the default service accounts in each namespace, and a Pod deployed by developers automatically inherits the image-pull-secret from the default service account in Pod's namespace.

This is done manually by following command for each Kubernetes namespace.

```
kubectl create secret docker-registry image-pull-secret \
  -n <your-namespace> \
  --docker-server=<your-registry-server> \
  --docker-username=<your-name> \
  --docker-password=<your-pword> \
  --docker-email=<your-email>

kubectl patch serviceaccount default \
  -p "{\"imagePullSecrets\": [{\"name\": \"image-pull-secret\"}]}" \
  -n <your-namespace>
```

And it could be automated with a simple program like imagepullsecret-patcher.

## Contribute

Development Environment

- Go 1.19
