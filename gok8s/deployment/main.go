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
	deploymentModel := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "image-classifier-resnet101",
			Namespace: "default",
			Labels:    map[string]string{"app": "image-classifier", "version": "resnet101"},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "image-classifier", "version": "resnet101"},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "image-classifier", "version": "resnet101"},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "tf-serving",
							Image:           "tensorflow/serving",
							Args:            []string{"--model_name=image_classifier", "--model_base_path=gs://clean-pen-305004-bucket/resnet_101"},
							ImagePullPolicy: v1.PullIfNotPresent,
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8501,
									Protocol:      v1.ProtocolTCP,
								},
								{
									Name:          "grpc",
									ContainerPort: 8500,
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
	_, err = clientset.AppsV1().Deployments("default").Create(ctx, deploymentModel, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Update other Model
	dpResNet50, err := clientset.AppsV1().Deployments("default").Get(ctx, "image-classifier-resnet50", metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	dpResNet50.Spec.Template.Spec.Containers[0].Args = []string{"--model_name=image_classifier", "--model_base_path=gs://clean-pen-305004-bucket/resnet_50"}
	_, err = clientset.AppsV1().Deployments("default").Update(ctx, dpResNet50, metav1.UpdateOptions{})
	if err != nil {
		panic(err.Error())
	}
}
