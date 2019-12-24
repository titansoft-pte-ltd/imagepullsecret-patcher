## deploy-example

Here is an example deployment to a kubernetes cluster.

Remember to change the Secret is specified in [2_deployment.yaml](deploy-example/kubernetes-manifest/2_deployment.yaml#L8). It's a base64-encoded json string which has credentials to the private registries.

To manually create such secret can follow https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/#create-a-secret-by-providing-credentials-on-the-command-line.

```
kubectl create secret docker-registry image-pull-secret \
  -n imagepullsecret-patcher \
  --docker-server=<your-registry-server> \
  --docker-username=<your-name> \
  --docker-password=<your-pword> \
  --docker-email=<your-email>
```
