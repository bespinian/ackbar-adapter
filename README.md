# ackbar-adapter

A custom metrics adapter which makes `ackbar`'s `partitionToWorkerRatio` available as an external metric in the metrics API of Kubernetes.

## Deployment

```shell
kubectl apply -f k8s
```

## Smoke test

Make a raw request to the Kubernetes metrics API.

```shell
kubectl get --raw "/apis/external.metrics.k8s.io/v1beta1/namespaces/default/partition-to-worker-ratio"
```

You should be seeing a metrics item per context in your `ackbar` instance.

```json
{
  "kind": "ExternalMetricValueList",
  "apiVersion": "external.metrics.k8s.io/v1beta1",
  "metadata": {},
  "items": [
    {
      "metricName": "partition-to-worker-ratio",
      "metricLabels": {
        "context": "3c0d1a1d-2e58-4032-a7dd-cf93a3eef4d3"
      },
      "timestamp": "2024-06-10T11:54:54Z",
      "value": "3"
    },
    {
      "metricName": "partition-to-worker-ratio",
      "metricLabels": {
        "context": "0d29016f-00a4-468d-a6b9-c03262cfca1c"
      },
      "timestamp": "2024-06-10T11:54:54Z",
      "value": "0"
    }
  ]
}
```
