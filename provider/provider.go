package provider

import (
	"context"
	"sync"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	"k8s.io/metrics/pkg/apis/external_metrics"

	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider/defaults"
)

type externalMetric struct {
	info   provider.ExternalMetricInfo
	labels map[string]string
	value  external_metrics.ExternalMetricValue
}

type ackbarProvider struct {
	//	defaults.DefaultCustomMetricsProvider
	defaults.DefaultExternalMetricsProvider
	client dynamic.Interface
	mapper apimeta.RESTMapper

	valuesLock      sync.RWMutex
	externalMetrics []externalMetric

	ackbarURL string
}

var testingExternalMetrics = []externalMetric{
	{
		info: provider.ExternalMetricInfo{
			Metric: "my-external-metric",
		},
		labels: map[string]string{"foo": "bar"},
		value: external_metrics.ExternalMetricValue{
			MetricName: "my-external-metric",
			MetricLabels: map[string]string{
				"foo": "bar",
			},
			Value: *resource.NewQuantity(42, resource.DecimalSI),
		},
	},
	{
		info: provider.ExternalMetricInfo{
			Metric: "my-external-metric",
		},
		labels: map[string]string{"foo": "baz"},
		value: external_metrics.ExternalMetricValue{
			MetricName: "my-external-metric",
			MetricLabels: map[string]string{
				"foo": "baz",
			},
			Value: *resource.NewQuantity(43, resource.DecimalSI),
		},
	},
	{
		info: provider.ExternalMetricInfo{
			Metric: "other-external-metric",
		},
		labels: map[string]string{},
		value: external_metrics.ExternalMetricValue{
			MetricName:   "other-external-metric",
			MetricLabels: map[string]string{},
			Value:        *resource.NewQuantity(44, resource.DecimalSI),
		},
	},
}

func NewProvider(client dynamic.Interface, mapper apimeta.RESTMapper, ackbarUrl string) provider.ExternalMetricsProvider {
	return &ackbarProvider{
		client:          client,
		mapper:          mapper,
		ackbarURL:       ackbarUrl,
		externalMetrics: testingExternalMetrics,
	}
}

func (p *ackbarProvider) GetExternalMetric(_ context.Context, _ string, metricSelector labels.Selector, info provider.ExternalMetricInfo) (*external_metrics.ExternalMetricValueList, error) {
	p.valuesLock.RLock()
	defer p.valuesLock.RUnlock()

	matchingMetrics := []external_metrics.ExternalMetricValue{}
	for _, metric := range p.externalMetrics {
		if metric.info.Metric == info.Metric &&
			metricSelector.Matches(labels.Set(metric.labels)) {
			metricValue := metric.value
			metricValue.Timestamp = metav1.Now()
			matchingMetrics = append(matchingMetrics, metricValue)
		}
	}
	return &external_metrics.ExternalMetricValueList{
		Items: matchingMetrics,
	}, nil
}
