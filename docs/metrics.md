## Metrics

The Local-pvc-releaser controller is publishing the base metrics that are provided by KubeBuilder + additional custom metric indicating about successful PVC deletion.
Those metrics are exposed by prometheus exporter.

The default metrics that are instrumented by default are:

* Total number of reconciliation errors per controller
* Length of reconcile queue per controller
* Reconciliation latency
* Usual resource metrics such as CPU, memory usage, file descriptor usage
* Go runtime metrics such as number of Go routines, GC duration

[Prometheus](https://prometheus.io/) is a standard way to represent metrics in a modern cross-platform manner. 
The exporter can expose metrics both on HTTP (8080) and on HTTPS (8443) using Kubernetes auth-proxy.

<br>

___
## Custom metrics
**`deleted_pvc`**

Labels: `namespace, controller_name, dryrun`
<br>
Description: The number of successful PVC objects that got deleted by the controller
