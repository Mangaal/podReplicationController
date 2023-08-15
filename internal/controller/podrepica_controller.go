/*
Copyright 2023.

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

package controller

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Mangaal/podReplicationController/api/v1alpha1"
	podreplicaappv1alpha1 "github.com/Mangaal/podReplicationController/api/v1alpha1"
)

// PodRepicaReconciler reconciles a PodRepica object
type PodRepicaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=podreplica-app.my.customecontroller,resources=podrepicas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=podreplica-app.my.customecontroller,resources=podrepicas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=podreplica-app.my.customecontroller,resources=podrepicas/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PodRepica object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *PodRepicaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// TODO(user): your logic here

	fmt.Println("New Request Received")

	podRepica := &v1alpha1.PodRepica{}

	err := r.Get(ctx, req.NamespacedName, podRepica)

	if err != nil {
		fmt.Println("Resource delted not found")

		return ctrl.Result{}, nil
	}

	loop := *podRepica.Spec.Replicas
	for loop > 0 {

		name := podRepica.Name + strconv.Itoa(loop)
		pod := &coreV1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: podRepica.Namespace,
			},

			Spec: podRepica.Spec.Template,
		}

		if err := controllerutil.SetControllerReference(podRepica, pod, r.Scheme); err != nil {
			fmt.Println("Error Seting  controllerReference :" + err.Error())
		}

		existingPod := &coreV1.Pod{}

		err := r.Get(ctx, client.ObjectKey{Name: pod.Name, Namespace: pod.Namespace}, existingPod)

		if err != nil {
			fmt.Println("Error Geting Pod  Or Pod Not Found")

			err := r.Create(ctx, pod)

			if err != nil {
				fmt.Println("Error Creating Pod")
			}

			fmt.Println("Pod " + pod.Name + " created")

		} else {

			if !reflect.DeepEqual(pod.Spec, existingPod.Spec) {

				for i := 0; len(pod.Spec.Containers) > i; i++ {

					existingPod.Spec.Containers[i].Image = pod.Spec.Containers[i].Image

				}

				err = r.Update(ctx, existingPod)

				if err != nil {
					fmt.Println("Error Updating Pod", err)
				}

				fmt.Println("Pod " + pod.Name + " Updated")

			}

		}

		loop--
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodRepicaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&podreplicaappv1alpha1.PodRepica{}).
		Owns(&coreV1.Pod{}).
		Complete(r)
}
