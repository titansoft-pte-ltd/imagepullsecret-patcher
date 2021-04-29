## Installation

```bash
$ helm repo add titansoft-pte-ltd titansoft-pte-ltd/imagepullsecret-patcher
$ helm search repo titansoft-pte-ltd
$ helm install my-release titansoft-pte-ltd/<chart>
```

helm install -n imagepullsecret-patcher --create-namespace imagepullsecret-patcher imagepullsecret-patcher --set dockerConfigJsonBase64Encoded="REPLACE_ME"