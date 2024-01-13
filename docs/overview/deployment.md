# Deployment

This section describes how to deploy `kraut`.

## Using `helm`

> **Note:** This is the recommended way to deploy `kraut`.

```bash
helm create namespace kraut-system
helm upgrade -i -n kraut-system kraut oci://ghcr.io/nicklasfrahm/kraut:v0.1.2
```

Below you may find some of the most commonly customized configuration options.

| Parameter                 | Description                       | Default |
| ------------------------- | --------------------------------- | ------- |
| `operator.logging.level`  | The log level of the operator.    | `info`  |
| `operator.logging.format` | May either be `pretty` or `json`. | `json`  |
