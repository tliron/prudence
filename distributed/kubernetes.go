package distributed

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func DiscoverKubernetesNodes() []string {
	var nodes []string

	namespace := "workspace"
	selector := "app.kubernetes.io/instance=prudence-hello-world"

	if config, err := rest.InClusterConfig(); err == nil {
		if client, err := kubernetes.NewForConfig(config); err == nil {
			if pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: selector}); err == nil {
				for _, pod := range pods.Items {
					nodes = append(nodes, pod.Status.PodIP)
				}
			} else { //if !errors.IsNotFound(err) {
				log.Errorf("%s", err)
			}
		} else {
			log.Errorf("%s", err)
		}
	} else {
		log.Errorf("%s", err)
	}

	return nodes
}
