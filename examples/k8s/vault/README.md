- Configuration

Check values.yaml inside in this path

- Install

```bash
helm dependency build
helm upgrade --install vault ./ -f ./values.yaml --create-namespace
```
