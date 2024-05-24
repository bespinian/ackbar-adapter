package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
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

type ackbarContexts []struct {
	ID                      string  `json:"id"`
	Name                    string  `json:"name"`
	LivenessIntervalSeconds int     `json:"livenessIntervalSeconds"`
	MaxPartitionsPerWorker  int     `json:"maxPartitionsPerWorker"`
	PartitionToWorkerRatio  float64 `json:"partitionToWorkerRatio"`
}

func NewProvider(client dynamic.Interface, mapper apimeta.RESTMapper, ackbarUrl string) provider.ExternalMetricsProvider {
	return &ackbarProvider{
		client:    client,
		mapper:    mapper,
		ackbarURL: ackbarUrl,
	}
}

func (p *ackbarProvider) GetExternalMetric(_ context.Context, _ string, metricSelector labels.Selector, info provider.ExternalMetricInfo) (*external_metrics.ExternalMetricValueList, error) {
	p.valuesLock.RLock()
	defer p.valuesLock.RUnlock()

	contextsUrl := fmt.Sprintf("%s/contexts", p.ackbarURL)
	resp, err := http.Get(contextsUrl)
	if err != nil {
		klog.Errorf("No response from %s: %v", p.ackbarURL, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	var contexts ackbarContexts
	if err := json.Unmarshal(body, &contexts); err != nil {
		klog.Errorf("Cannot unmarshal JSON: %v", err)
	}

	contextMetrics := []externalMetric{}

	for _, context := range contexts {
		partitionToWorkerRationString := strconv.FormatFloat(context.PartitionToWorkerRatio, 'f', -1, 64)
		metricValue, err := resource.ParseQuantity(partitionToWorkerRationString)
		if err != nil {
			klog.Errorf("Cannot convert ratio to quantity: %v", err)
		}
		contextMetrics = append(contextMetrics, externalMetric{
			info: provider.ExternalMetricInfo{
				Metric: "partition-to-worker-ratio",
			},
			labels: map[string]string{
				"context": context.ID,
			},
			value: external_metrics.ExternalMetricValue{
				MetricName: "partition-to-worker-ratio",
				MetricLabels: map[string]string{
					"context": context.ID,
				},
				Value: metricValue,
			},
		})
	}

	matchingMetrics := []external_metrics.ExternalMetricValue{}
	for _, metric := range contextMetrics {
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
