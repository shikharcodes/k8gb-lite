package controllers

/*
Copyright 2022 The k8gb Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Generated by GoLic, for more details see: https://github.com/AbsaOSS/golic
*/

import (
	"context"

	"cloud.example.com/annotation-operator/controllers/reconciliation"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	externaldns "sigs.k8s.io/external-dns/endpoint"
)

// SetupWithManager sets up the controller with the Manager.
func (r *AnnoReconciler) SetupWithManager(mgr ctrl.Manager) error {

	ingressHandler := handler.EnqueueRequestsFromMapFunc(
		func(a client.Object) []reconcile.Request {
			rs1, err := reconciliation.NewLoopState(a.(*netv1.Ingress))
			if err != nil {
				return nil
			}
			rs2, result, _ := r.IngressMapper.Get(rs1.NamespacedName)
			switch result {
			case reconciliation.MapperResultExists:
				if !r.IngressMapper.Equal(rs1, rs2) {
					return []reconcile.Request{{NamespacedName: rs1.NamespacedName}}
				}
			default:
				return nil
			}
			return nil
		})

	serviceEndpointHandler := handler.EnqueueRequestsFromMapFunc(
		func(a client.Object) []reconcile.Request {
			ingList := &netv1.IngressList{}
			c := mgr.GetClient()
			err := c.List(context.TODO(), ingList, client.InNamespace(a.GetNamespace()))
			if err != nil {
				r.Log.Info().Msg("Can't fetch ingress objects")
				return nil
			}
			for _, ing := range ingList.Items {
				for _, rule := range ing.Spec.Rules {
					for _, path := range rule.HTTP.Paths {
						if path.Backend.Service != nil && path.Backend.Service.Name == a.GetName() {
							return []reconcile.Request{{NamespacedName: types.NamespacedName{Namespace: a.GetNamespace(), Name: ing.Name}}}
						}
					}
				}
			}
			return nil
		})

	return ctrl.NewControllerManagedBy(mgr).
		For(&netv1.Ingress{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&externaldns.DNSEndpoint{}).
		Watches(&source.Kind{Type: &netv1.Ingress{}}, ingressHandler).
		Watches(&source.Kind{Type: &corev1.Endpoints{}}, serviceEndpointHandler, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}
