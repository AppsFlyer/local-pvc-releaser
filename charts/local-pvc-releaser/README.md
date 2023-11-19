# Local-PVC-Releaser Helm Chart
___

A Helm chart for local-pvc-releaser, a PVC controller project built for Kubernetes.

## Installing the Chart

```console
$ helm repo add local-pvc-releaser https://AppsFlyer.github.io/local-pvc-releaser
$ helm install -n <namespace> <release-name> local-pvc-releaser/local-pvc-rele0.1aser
```

## Uninstalling the Chart

To uninstall/delete the `local-pvc-releaser` deployment:

```console
$ helm delete --purge local-pvc-releaser
```


## Configuring the chart

The following table lists the configurable parameters of the Local PVC Releaser for Kubernetes chart and their
default values.

| Parameter                                                | Description                                               | Default                                                          |
|----------------------------------------------------------|-----------------------------------------------------------|------------------------------------------------------------------|
| `controller.name`                                        | Number of replicas for the controller                     | `local-pvc-releaser`                                             |
| `controller.replicas`                                    | Name of the controller pod                                | `1`                                                              |
| `controller,image.repository`                            | Local PVC releaser repository name                        | `appsflyer/local-pvc-releaser` |
| `controller.image.tag`                                   | Local PVC releaser image tag                              | `v0.1.0`                                                         |
| `controller.image.pullPolicy`                            | Image pull policy                                         | `Always`                                                         |
| `controller.resources`                                   | Local PVC Releaser resource requests & limits             | `{}`                                                             |
| `controller.pvcAnnotationSelector.enabled`               | Enable PVC Annotation selector                            | `true`                                                           |
| `controller.pvcAnnotationSelector.customAnnotationKey`   | Custom PVC Annotation filter key                          | `appsflyer.com/local-pvc-releaser`                               |
| `controller.pvcAnnotationSelector.customAnnotationValue` | Custom PVC Annotation filter value                        | `enabled`                                                        |
| `controller.additionalAnnotations`                       | Additional annotations to be added to the deployment      | `{}`                                                             |
| `controller.additionalLabels`                            | Additional labels to be added to the deployment           | `{}`                                                             |
| `controller.tolerations`                                 | Node taints to tolerate                                   | `[]`                                                             |
| `controller.affinity`                                    | Pod affinity                                              | `{}`                                                             |
| `controller.dryRun`                                      | Enable the controller in dry-run mode                     | `false`                                                          |
| `controller.logLevel`                                    | Define controller log level                               | `info`                                                           |
| `controller.loggingDevMode`                              | Enable the controller logger with stack tracing           | `false`                                                          |
| `controller.extraEnv`                                    | Extra environment variables to be added to the deployment | `{}`                                                             |
| `prometheus.enabled`                                     | Enabling prometheus exporter                              | `true`                                                           |
| `prometheus.portType`                                    | Exposing the metrics endpoint on http/https port          | `http`                                                           |


## Dependencies

In order to use the ServiceMonitor CRD (disabled by default), you must have [Prometheus-Operator](https://github.com/prometheus-operator/prometheus-operator) installed on your cluster.