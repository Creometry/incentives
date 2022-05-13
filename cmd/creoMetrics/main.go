package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// PodMetricsList : PodMetricsList
type PodMetricsList struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
		SelfLink string `json:"selfLink"`
	} `json:"metadata"`
	Items []struct {
		Metadata struct {
			Name              string    `json:"name"`
			Namespace         string    `json:"namespace"`
			SelfLink          string    `json:"selfLink"`
			CreationTimestamp time.Time `json:"creationTimestamp"`
		} `json:"metadata"`
		Timestamp  time.Time `json:"timestamp"`
		Window     string    `json:"window"`
		Containers []struct {
			Name  string `json:"name"`
			Usage struct {
				CPU    string `json:"cpu"`
				Memory string `json:"memory"`
			} `json:"usage"`
		} `json:"containers"`
	} `json:"items"`
}

func getMetrics(clientset *kubernetes.Clientset, pods *PodMetricsList) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	data, err := clientset.RESTClient().Get().AbsPath("apis/metrics.k8s.io/v1beta1/pods").DoRaw(ctx)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &pods)
	return err
}

func main() {
	// creates the in-cluster config
	// https://github.com/kubernetes/client-go/tree/master/examples#configuration
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// dirname, err := os.UserHomeDir()
	// kubeconfig := dirname + "/.kube/config"
	// fmt.Println(kubeconfig)
	// // config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	// config, err := clientcmd.BuildConfigFromFlags("https://10.96.0.1:6443", kubeconfig)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	var pods PodMetricsList
	err = getMetrics(clientset, &pods)
	if err != nil {
		panic(err.Error())
	}
	for _, m := range pods.Items {
		fmt.Println(m.Metadata.Name, m.Metadata.Namespace, m.Timestamp.String())
	}
}
