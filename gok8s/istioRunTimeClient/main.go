// https://istio.io/latest/blog/2019/announcing-istio-client-go/
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	inetworkingv1alpha3 "istio.io/api/networking/v1alpha3"
	iv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	istioScheme "istio.io/client-go/pkg/clientset/versioned/scheme"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

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

	// get configigs.k8s.io/controller-runtime/pkg/cl
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// Create runtime client
	crScheme := runtime.NewScheme()

	/*
		Schema:
		Golang type -> GVK(Group Version KInd)
	*/
	// istio
	if err = istioScheme.AddToScheme(crScheme); err != nil {
		fmt.Println(err)
	}

	// kubernetes native objects (service)
	if err = clientgoscheme.AddToScheme(crScheme); err != nil {
		fmt.Println(err)
	}

	// client set
	cl, err := runtimeclient.New(config, runtimeclient.Options{
		Scheme: crScheme,
	})
	if err != nil {
		panic(err.Error())
	}

	// get service
	svcget := &v1.Service{}
	err = cl.Get(ctx, types.NamespacedName{Namespace: "default", Name: "image-classifier"}, svcget)
	if err != nil {
		fmt.Println(err)
	}
	// change to clusterIP
	svcget.Spec.Type = v1.ServiceTypeClusterIP
	//update
	err = cl.Update(ctx, svcget)
	if err != nil {
		fmt.Println(err)
	}

	// create istio object
	vs := &iv1alpha3.VirtualService{
		TypeMeta: metav1.TypeMeta{
			Kind: "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "image-classifier",
			Namespace: "default",
		},
		Spec: inetworkingv1alpha3.VirtualService{
			Hosts:    []string{"*"},
			Gateways: []string{"image-classifier-gateway"},
			Http: []*inetworkingv1alpha3.HTTPRoute{
				{
					Route: []*inetworkingv1alpha3.HTTPRouteDestination{
						{
							Destination: &inetworkingv1alpha3.Destination{
								Host: "image-classifier",
								Port: &inetworkingv1alpha3.PortSelector{
									Number: 8501,
								},
								Subset: "resnet50",
							},
							Weight: 50,
						},
						{
							Destination: &inetworkingv1alpha3.Destination{
								Host: "image-classifier",
								Port: &inetworkingv1alpha3.PortSelector{
									Number: 8501,
								},
								Subset: "resnet101",
							},
							Weight: 50,
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
