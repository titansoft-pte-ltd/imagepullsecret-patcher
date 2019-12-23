## deploy-example

Here is an example deployment to an kubernetes cluster.

Note that the docker registry credential is specified in [2_deployment.yaml](deploy-example/kubernetes-manifest/2_deployment.yaml) file and stored in a secret. To manually create such secret can see https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/#create-a-secret-by-providing-credentials-on-the-command-line.