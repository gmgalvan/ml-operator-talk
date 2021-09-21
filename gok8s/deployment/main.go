package main

import (
	"context"
	"flag"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	ctx := context.Background()
	// get kubeconfig
	myDefaultKbConfig := os.Getenv("PATH_KUBECONFIG") // /home/user/.kube/config
	kubeconfig := flag.String("kubeconfig", myDefaultKbConfig, "kubeconfig file")
	flag.Parse()

	// obtain rest config
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// pass the rest config and obtain the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Define our deployment object
	replicas := int32(2)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hello-app",
			Namespace: "example",
			Labels:    map[string]string{"app": "hello-app"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "hello-app"},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "hello-app"},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "hello-app",
							Image: "docker7gm/hello-world-app:v1.0.0",
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 8080,
									Protocol:      v1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}
	// Create our deployment (/apis/apps/v1/deployments)
	_, err = clientset.AppsV1().Deployments("example").Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}
}
