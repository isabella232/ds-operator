/*
	Copyright 2020 ForgeRock AS.
*/

package controllers

import (
	"context"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	k8slog "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *DirectoryServiceReconciler) reconcileService(ctx context.Context, ds *directoryv1alpha1.DirectoryService, svcName string) (v1.Service, error) {
	log := k8slog.FromContext(ctx)
	// create or update the service
	var svc v1.Service
	svc.Name = svcName
	svc.Namespace = ds.Namespace

	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &svc, func() error {
		log.V(8).Info("CreateorUpdate service", "svc", svc)

		var err error
		// does the service not exist yet?
		if svc.CreationTimestamp.IsZero() {
			err = createService(ds, &svc)
			log.V(8).Info("Setting ownerref for service", "svc", svc.Name)
			_ = controllerutil.SetControllerReference(ds, &svc, r.Scheme)
		} else {
			// If the service exists already - we want to update any fields to bring its state into
			// alignment with the Custom Resource
			//err = updateService(&ds, &sts)
			log.V(8).Info("TODO: Handle update of ds service")
		}

		log.V(8).Info("svc after update/create", "svc", svc)
		return err
	})
	return svc, err

}

// Create the service for ds
func createService(ds *directoryv1alpha1.DirectoryService, svc *v1.Service) error {
	svcTemplate := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      createLabels(ds.Name, nil),
			Annotations: make(map[string]string),
			Name:        svc.Name,
			Namespace:   ds.Namespace,
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None", // headless service
			Selector: map[string]string{
				"app.kubernetes.io/name":     LabelApplicationName,
				"app.kubernetes.io/instance": ds.Name,
			},
			Ports: []v1.ServicePort{
				{
					Name: "tcp-admin",
					Port: 4444,
				},
				{
					Name: "tcp-ldap",
					Port: 1389,
				},
				{
					Name: "tcp-ldaps",
					Port: 1636,
				},
				{
					Name: "tcp-replication",
					Port: 8989,
				},
				{
					Name: "http",
					Port: 8080,
				},
			},
		},
	}

	svcTemplate.DeepCopyInto(svc)
	return nil // todo: can this ever fail?
}
