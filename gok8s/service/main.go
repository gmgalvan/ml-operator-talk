package main

import (
	"context"
	"flag"
	"os"

	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr" //target port

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	ctx := context.Background()
	// get kubeconfig
	myDefaultKbConfig := os.Getenv("PATH_KUBECONFIG") // /home/user/.kubeconfig
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

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hello-app-service",
			Namespace: "example",
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeLoadBalancer,
			Selector: map[string]string{"app": "hello-app"},
			Ports: []v1.ServicePort{
				{
					Protocol:   v1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.IntOrString{IntVal: 8080},
				},
			},
		},
	}

	// create a service with loadbalancer in example namespace (/api/v1/services)
	_, err = clientset.CoreV1().Services("example").Create(ctx, svc, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}

}
