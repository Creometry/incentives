package main

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

func main() {
	// creates the in-cluster config
	// https://github.com/kubernetes/client-go/tree/master/examples#configuration
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// this loop is used to collect metrics evrey 10 seconds

	for {
		// creates the clientset
		// clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		mc, err := metrics.NewForConfig(config)
		if err != nil {
			fmt.Println("Error", err)
			panic(err)
		}

		podMetrics, err := mc.MetricsV1beta1().PodMetricses(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("podMetric data received")

		for _, podMetric := range podMetrics.Items {
			podContainers := podMetric.Containers
			for _, container := range podContainers {
				cpuQuantity, ok := container.Usage.Cpu().AsInt64()
				memQuantity, ok := container.Usage.Memory().AsInt64()
				if !ok {
					return
				}
				msg := fmt.Sprintf("Container Name: %s \n CPU usage: %d \n Memory usage: %d", container.Name, cpuQuantity, memQuantity)
				fmt.Println(msg)
			}

		}

		time.Sleep(10 * time.Second)

	}

}
