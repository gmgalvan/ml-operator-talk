/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	mlappsv1alpha1 "github.com/ml-operator-talk/tf-canary-operator/api/v1alpha1"
	istionetworkv1alpha3 "istio.io/api/networking/v1alpha3"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// TfCanaryReconciler reconciles a TFCanary object
type TfCanaryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mlapps.demo.go,resources=tfcanaries,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mlapps.demo.go,resources=tfcanaries/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mlapps.demo.go,resources=tfcanaries/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=gateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=destinationrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TFCanary object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *TfCanaryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	// Fetch TFCanary
	tfCanaryDeploy := &mlappsv1alpha1.TfCanary{}
	err := r.Get(ctx, req.NamespacedName, tfCanaryDeploy)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("TfCanary resource not found. Ignoring since object must be delete")
			return ctrl.Result{}, nil
		}
	}

	// Check if Deployments already exists
	for _, model := range tfCanaryDeploy.Spec.Models {
		// Check if the deployment already exists, if not create a new one
		found := &appsv1.Deployment{}
		deployedModelName := tfCanaryDeploy.Name + "-" + model.Name
		err = r.Get(ctx, types.NamespacedName{Name: deployedModelName, Namespace: tfCanaryDeploy.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			// Define a new deployment
			dep := r.modelDeployment(ctx, *tfCanaryDeploy, model)
			log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			err = r.Create(ctx, dep)
			if err != nil {
				log.Error(err, "Failed to create new Model Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
				return ctrl.Result{}, err
			}
			//return ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			log.Error(err, "Failed to get Deployment")
			return ctrl.Result{}, err
		}
	}

	// Check if Services already exists
	svcFound := &v1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: tfCanaryDeploy.Name, Namespace: tfCanaryDeploy.Namespace}, svcFound)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Service
		svc := r.modelService(ctx, *tfCanaryDeploy)
		log.Info("Creating a new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		err = r.Create(ctx, svc)
		if err != nil {
			log.Error(err, "Failed to create new Model Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return ctrl.Result{}, err
		}
		// Service created successfully - return and requeue
		//return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}

	// Check if Gateway already exists
	gatewayFound := &istiov1alpha3.Gateway{}
	gatewayName := tfCanaryDeploy.Name + "-gateway"
	err = r.Get(ctx, types.NamespacedName{Name: gatewayName, Namespace: tfCanaryDeploy.Namespace}, gatewayFound)
	if err != nil && errors.IsNotFound(err) {
		// Define a new gateway
		gtw := r.ensureGateway(*tfCanaryDeploy)
		log.Info("Creating a new gateway", "gateway.Namespace", gtw.Namespace, "gateway.Name", gtw.Name)
		err = r.Create(ctx, gtw)
		if err != nil {
			log.Error(err, "Failed to create new gateway", "gateway.Namespace", gtw.Namespace, "gateway.Name", gtw.Name)
			return ctrl.Result{}, err
		}
		// gateway created successfully - return and requeue
		//return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Gateway")
		return ctrl.Result{}, err
	}

	// Check for istio destination rule
	destinationRulesFound := &istiov1alpha3.DestinationRule{}
	err = r.Get(ctx, types.NamespacedName{Name: tfCanaryDeploy.Name, Namespace: tfCanaryDeploy.Namespace}, destinationRulesFound)
	if err != nil && errors.IsNotFound(err) {
		// Define a new destination rule
		dr := r.createDestinationRules(*tfCanaryDeploy)
		log.Info("Creating a new destination rule", "destinationRule.Namespace", dr.Namespace, "destinationRule.Name", dr.Name)
		err = r.Create(ctx, dr)
		if err != nil {
			log.Error(err, "Failed to create new destination rule", "destinationRule.Namespace", dr.Namespace, "destinationRule.Name", dr.Name)
			return ctrl.Result{}, err
		}
		// destination rule created successfully - return and requeue
		//return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get destination rule")
		return ctrl.Result{}, err
	}

	// check for virtual service config
	virtualServiceFound := &istiov1alpha3.VirtualService{}
	err = r.Get(ctx, types.NamespacedName{Name: tfCanaryDeploy.Name, Namespace: tfCanaryDeploy.Namespace}, virtualServiceFound)
	if err != nil && errors.IsNotFound(err) {
		// Define a new virtual service
		vs := r.createWeightedVirtualService(*tfCanaryDeploy)
		log.Info("Creating a new virtual service", "virtualService.Namespace", vs.Namespace, "virtualService.Name", vs.Name)
		err = r.Create(ctx, vs)
		if err != nil {
			log.Error(err, "Failed to create new virtual service", "virtualService.Namespace", vs.Namespace, "virtualService.Name", vs.Name)
			return ctrl.Result{}, err
		}
		// virtual service created successfully - return and requeue
		//return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get virtual service")
		return ctrl.Result{}, err
	}
	// update
	deploymentsFound := &appsv1.DeploymentList{}
	// set a list of options from metav1
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"app": tfCanaryDeploy.Name}}
	err = r.Client.List(ctx, deploymentsFound, &client.ListOptions{Namespace: tfCanaryDeploy.Namespace, LabelSelector: labels.Set(labelSelector.MatchLabels).AsSelector()})
	if err != nil {
		log.Info("an error occours while trying to get deployments")
	}
	for idx := range deploymentsFound.Items {
		deploymentFound := deploymentsFound.Items[idx]
		foundModelBaseLoc := strings.Replace(deploymentFound.Spec.Template.Spec.Containers[0].Args[1], "--model_base_path=", "", -1)
		modelLoc := tfCanaryDeploy.Spec.Models[idx].Location
		if foundModelBaseLoc != modelLoc {
			newModelLoc := fmt.Sprintf("--model_base_path=%v", modelLoc)
			deploymentFound.Spec.Template.Spec.Containers[0].Args[1] = newModelLoc
			err = r.Update(ctx, &deploymentFound)
			if err != nil {
				log.Error(err, "Failed to update Deployment", "Deployment.Namespace", tfCanaryDeploy.Namespace, "Deployment.Name", tfCanaryDeploy.Name)
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: time.Minute * 2}, nil
		}
		if tfCanaryDeploy.Spec.Models[idx].Weight != virtualServiceFound.Spec.Http[0].Route[idx].Weight {
			// update all in order to ensure that sums 100
			for j, _ := range virtualServiceFound.Spec.Http[0].Route {
				virtualServiceFound.Spec.Http[0].Route[j].Weight = tfCanaryDeploy.Spec.Models[j].Weight
			}
			err = r.Update(ctx, virtualServiceFound)
			if err != nil {
				log.Error(err, "Failed to update weight", "Deployment.Namespace", tfCanaryDeploy.Namespace, "Deployment.Name", tfCanaryDeploy.Name)
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *TfCanaryReconciler) createWeightedVirtualService(tfCanrayDep mlappsv1alpha1.TfCanary) *istiov1alpha3.VirtualService {
	gatewayName := tfCanrayDep.Name + "-gateway"
	routeDestination := []*istionetworkv1alpha3.HTTPRouteDestination{}
	for _, model := range tfCanrayDep.Spec.Models {
		routeDestination = append(routeDestination, &istionetworkv1alpha3.HTTPRouteDestination{
			Destination: &istionetworkv1alpha3.Destination{
				Host: tfCanrayDep.Name,
				Port: &istionetworkv1alpha3.PortSelector{
					Number: 8501,
				},
				Subset: model.Name,
			},
			Weight: model.Weight,
		})
	}
	vs := &istiov1alpha3.VirtualService{
		TypeMeta: metav1.TypeMeta{
			Kind: "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      tfCanrayDep.Name,
			Namespace: tfCanrayDep.Namespace,
		},
		Spec: istionetworkv1alpha3.VirtualService{
			Hosts:    []string{"*"},
			Gateways: []string{gatewayName},
			Http: []*istionetworkv1alpha3.HTTPRoute{
				{
					Route: routeDestination,
				},
			},
		},
	}
	ctrl.SetControllerReference(&tfCanrayDep, vs, r.Scheme)
	return vs
}

func (r *TfCanaryReconciler) createDestinationRules(tfCanrayDep mlappsv1alpha1.TfCanary) *istiov1alpha3.DestinationRule {
	subsets := []*istionetworkv1alpha3.Subset{}
	for _, model := range tfCanrayDep.Spec.Models {
		subsets = append(subsets, &istionetworkv1alpha3.Subset{
			Name:   model.Name,
			Labels: map[string]string{"version": model.Name},
		})
	}
	dr := &istiov1alpha3.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			Kind: "DestinationRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      tfCanrayDep.Name,
			Namespace: tfCanrayDep.Namespace,
		},
		Spec: istionetworkv1alpha3.DestinationRule{
			Host:    tfCanrayDep.Name,
			Subsets: subsets,
		},
	}
	ctrl.SetControllerReference(&tfCanrayDep, dr, r.Scheme)
	return dr
}

func (r *TfCanaryReconciler) ensureGateway(tfCanrayDep mlappsv1alpha1.TfCanary) *istiov1alpha3.Gateway {
	gateway := &istiov1alpha3.Gateway{
		TypeMeta: metav1.TypeMeta{
			Kind: "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      tfCanrayDep.Name + "-gateway",
			Namespace: "default",
		},
		Spec: istionetworkv1alpha3.Gateway{
			Selector: map[string]string{"istio": "ingressgateway"},
			Servers: []*istionetworkv1alpha3.Server{
				{
					Port: &istionetworkv1alpha3.Port{
						Number:   80,
						Name:     "http",
						Protocol: "HTTP",
					},
					Hosts: []string{"*"},
				},
			},
		},
	}
	ctrl.SetControllerReference(&tfCanrayDep, gateway, r.Scheme)
	return gateway
}

func (r *TfCanaryReconciler) modelDeployment(ctx context.Context, tfCanrayDep mlappsv1alpha1.TfCanary, model mlappsv1alpha1.Model) *appsv1.Deployment {
	modelBasePath := fmt.Sprintf("--model_base_path=%v", model.Location)
	modelName := fmt.Sprintf("--model_name=%v", tfCanrayDep.Name)
	deploymentModel := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tfCanrayDep.Name + "-" + model.Name,
			Namespace: tfCanrayDep.Namespace,
			Labels:    map[string]string{"app": tfCanrayDep.Name, "version": model.Name},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": tfCanrayDep.Name, "version": model.Name},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": tfCanrayDep.Name, "version": model.Name},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "tf-serving",
							Image:           "tensorflow/serving",
							Args:            []string{modelName, modelBasePath},
							ImagePullPolicy: v1.PullIfNotPresent,
							ReadinessProbe: &v1.Probe{
								Handler:             v1.Handler{TCPSocket: &v1.TCPSocketAction{Port: intstr.IntOrString{IntVal: 8500}}},
								InitialDelaySeconds: 10,
								PeriodSeconds:       5,
								FailureThreshold:    10,
							},
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
	ctrl.SetControllerReference(&tfCanrayDep, deploymentModel, r.Scheme)
	return deploymentModel
}

func (r *TfCanaryReconciler) modelService(ctx context.Context, tfCanrayDep mlappsv1alpha1.TfCanary) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tfCanrayDep.Name,
			Namespace: tfCanrayDep.Namespace,
			Labels:    map[string]string{"app": tfCanrayDep.Name, "service": tfCanrayDep.Name},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				{
					Port:     8500,
					Protocol: v1.ProtocolTCP,
					Name:     "tf-serving-grpc",
				},
				{
					Port:     8501,
					Protocol: v1.ProtocolTCP,
					Name:     "tf-serving-http",
				},
			},
			Selector: map[string]string{"app": tfCanrayDep.Name},
		},
	}
	ctrl.SetControllerReference(&tfCanrayDep, svc, r.Scheme)
	return svc
}

// SetupWithManager sets up the controller with the Manager.
func (r *TfCanaryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mlappsv1alpha1.TfCanary{}).
		Complete(r)
}
