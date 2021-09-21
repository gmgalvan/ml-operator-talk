// https://istio.io/latest/blog/2019/announcing-istio-client-go/
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	inetworkingv1alpha3 "istio.io/api/networking/v1alpha3"
	iv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istio "istio.io/client-go/pkg/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	ctx := context.Background()

	// get the config rest
	myDefaultKbConfig := os.Getenv("PATH_KUBECONFIG")
	kubeconfig := flag.String("kubeconfig", myDefaultKbConfig, "kubeconfig file")
	flag.Parse()

	// get config
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// Create runtime client
	crScheme := runtime.NewScheme()

	// Add the run time client istio
	if err = istio.AddToScheme(crScheme); err != nil {
		fmt.Println(err)
	}

	cl, err := runtimeclient.New(config, runtimeclient.Options{
		Scheme: crScheme,
	})
	if err != nil {
		panic(err.Error())
	}

	// create istio object
	vs := &iv1alpha3.VirtualService{
		TypeMeta: metav1.TypeMeta{
			Kind: "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hello-app",
			Namespace: "default",
		},
		Spec: inetworkingv1alpha3.VirtualService{
			Hosts:    []string{"*"},
			Gateways: []string{"hello-app"},
			Http: []*inetworkingv1alpha3.HTTPRoute{
				{
					Route: []*inetworkingv1alpha3.HTTPRouteDestination{
						{
							Destination: &inetworkingv1alpha3.Destination{
								Host: "hello-app",
								Port: &inetworkingv1alpha3.PortSelector{
									Number: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	// crating the istio vistual service
	err = cl.Create(ctx, vs)
	if err != nil {
		fmt.Println(err)
	}

}
