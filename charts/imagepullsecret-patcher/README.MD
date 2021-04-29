## Installation

```bash
$ helm repo add titansoft-pte-ltd titansoft-pte-ltd/imagepullsecret-patcher
$ helm search repo titansoft-pte-ltd
$ helm install -n imagepullsecret-patcher --create-namespace imagepullsecret-patcher titansoft-pte-ltd/imagepullsecret-patcher --set dockerConfigJsonBase64Encoded="REPLACE_ME"
```