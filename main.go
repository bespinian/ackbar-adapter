package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"

	basecmd "sigs.k8s.io/custom-metrics-apiserver/pkg/cmd"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"

	// make this the path to the provider that you just wrote
	yourprov "bespinian.io/ackbar-adapter/provider"
)

type YourAdapter struct {
	basecmd.AdapterBase

	// the message printed on startup
	Message string

	// The URL for the ackbar instance to use for metrics
	AckbarURL string
}

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	// initialize the flags, with one custom flag for the message and one for the URL of ackbar
	cmd := &YourAdapter{}
	cmd.Flags().StringVar(&cmd.Message, "msg", "starting adapter...", "startup message")
	cmd.Flags().StringVar(&cmd.AckbarURL, "ackbar-url", "", "The URL of the ackbar instance to use for metrics")

	// make sure you get the klog flags
	logs.AddGoFlags(flag.CommandLine)
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	cmd.Flags().Parse(os.Args)

	if cmd.AckbarURL == "" {
		klog.Fatalf("ackbar URL not configured. Please set command line flag --ackbar-url")
	}

	provider := cmd.makeProviderOrDie()
	// cmd.WithCustomMetrics(provider)
	cmd.WithExternalMetrics(provider)

	klog.Infof("ackbar URL configured as %s", cmd.AckbarURL)
	klog.Infof(cmd.Message)
	if err := cmd.Run(wait.NeverStop); err != nil {
		klog.Fatalf("unable to run custom metrics adapter: %v", err)
	}
}

func (a *YourAdapter) makeProviderOrDie() provider.ExternalMetricsProvider {
	client, err := a.DynamicClient()
	if err != nil {
		klog.Fatalf("unable to construct dynamic client: %v", err)
	}

	mapper, err := a.RESTMapper()
	if err != nil {
		klog.Fatalf("unable to construct discovery REST mapper: %v", err)
	}

	return yourprov.NewProvider(client, mapper, a.AckbarURL)
}
