package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/env"
	"log"
	"net/http"
	"time"
)

var (
	K8sLivenessProbes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Help: "todo",
		Name: "k8s_liveness_probe_exists",
	}, []string{"namespace", "deployment", "container"})
	K8sReadinessProbes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Help: "todo",
		Name: "k8s_readiness_probe_exists",
	}, []string{"namespace", "deployment", "container"})
)

func Init() {
	prometheus.MustRegister(K8sLivenessProbes)
	prometheus.MustRegister(K8sReadinessProbes)
}

func GetK8sClient() (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", env.GetString("KUBECONFIG", ""))
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func main() {
	Init()

	ctx := context.Background()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8080", nil)
	}()

	for {
		client, err := GetK8sClient()
		if err != nil {
			log.Fatal(err)
		}
		deployments, err := client.AppsV1().Deployments("").List(ctx, v1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}

		for _, deploy := range deployments.Items {
			for _, container := range deploy.Spec.Template.Spec.Containers {
				var livenessExists float64 = 0.0
				var readinessExists float64 = 0.0

				if container.LivenessProbe != nil {
					livenessExists = 1.0
				}
				if container.ReadinessProbe != nil {
					readinessExists = 1.0
				}

				K8sLivenessProbes.With(prometheus.Labels{"namespace": deploy.Namespace, "deployment": deploy.Name, "container": container.Name}).Set(livenessExists)
				K8sReadinessProbes.With(prometheus.Labels{"namespace": deploy.Namespace, "deployment": deploy.Name, "container": container.Name}).Set(readinessExists)
			}
		}

		time.Sleep(time.Second * 60)
	}
}
