/*
Copyright 2025.

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
	"encoding/base64"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	k8sstartkubernetescomv2 "k8s.startkubernetes.com/v2/api/v2"
)

// PdfDocumentReconciler reconciles a PdfDocument object
type PdfDocumentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=k8s.startkubernetes.com.my.domain,resources=pdfdocuments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=k8s.startkubernetes.com.my.domain,resources=pdfdocuments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=k8s.startkubernetes.com.my.domain,resources=pdfdocuments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PdfDocument object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *PdfDocumentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Create a logger instance from the context
	logger := logf.FromContext(ctx)

	// Fetch the PdfDocument resource from the Kubernetes cluster
	var pdfDoc k8sstartkubernetescomv2.PdfDocument
	if err := r.Get(ctx, req.NamespacedName, &pdfDoc); err != nil {
		if apierrors.IsNotFound(err) {
			// Log if the PdfDocument resource is deleted
			logger.Info("PdfDocument resource deleted", "name", req.Name, "namespace", req.Namespace)

			// Attempt to delete the associated Job
			jobName := req.Name + "-job"
			var job batchv1.Job
			err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: req.Namespace}, &job)
			if err == nil {
				if err := r.Delete(ctx, &job, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil {
					logger.Error(err, "failed to delete associated Job", "job", jobName)
					return ctrl.Result{}, err
				}
				logger.Info("Associated Job deleted successfully", "job", jobName)
			} else if !apierrors.IsNotFound(err) {
				logger.Error(err, "failed to fetch associated Job", "job", jobName)
				return ctrl.Result{}, err
			}

			// Return without requeuing if the resource is not found
			return ctrl.Result{}, nil
		} else {
			// Log an error if the PdfDocument resource cannot be fetched
			logger.Error(err, "unable to fetch PdfDocument")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	// Create a Job specification based on the PdfDocument resource
	jobSpec, err := r.createJob(pdfDoc)
	if err != nil {
		// Log an error if the Job specification cannot be created
		logger.Error(err, "unable to create job spec")
		// Return without requeuing if there is an error
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var existingJob batchv1.Job
	err = r.Get(ctx, types.NamespacedName{Name: jobSpec.Name, Namespace: jobSpec.Namespace}, &existingJob)
	if err == nil {
		// Job already exists, check if it succeeded
		if existingJob.Status.Succeeded > 0 {
			logger.Info("PDF generation job succeeded", "job", jobSpec.Name)
		} else {
			logger.Info("Job already exists, skipping creation", "job", jobSpec.Name)
		}
		return ctrl.Result{}, nil
	} else if !apierrors.IsNotFound(err) {
		// Some other error occurred
		logger.Error(err, "failed to check if Job exists")
		return ctrl.Result{}, err
	}

	// Create the Job resource in the Kubernetes cluster
	if err := r.Create(ctx, &jobSpec); err != nil {
		// Log an error if the Job resource cannot be created
		logger.Error(err, "unable to create job")
		return ctrl.Result{}, err
	}

	// Log that the Job was successfully created
	logger.Info("Job created successfully", "job", jobSpec.Name)

	// Return successfully without requeuing
	return ctrl.Result{}, nil
}

func (r *PdfDocumentReconciler) createJob(pdfDoc k8sstartkubernetescomv2.PdfDocument) (batchv1.Job, error) {
	image := "knsit/pandoc"
	base64text := base64.StdEncoding.EncodeToString([]byte(pdfDoc.Spec.Text))

	j := batchv1.Job{
		TypeMeta: metav1.TypeMeta{APIVersion: batchv1.SchemeGroupVersion.String(), Kind: "Job"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pdfDoc.Name + "-job",
			Namespace: pdfDoc.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyOnFailure,
					InitContainers: []corev1.Container{
						{
							Name:    "store-to-md",
							Image:   "alpine",
							Command: []string{"/bin/sh"},
							Args:    []string{"-c", fmt.Sprintf("echo %s | base64  -d >> /data/text.md", base64text)},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data-volume",
									MountPath: "/data",
								},
							},
						},
						{
							Name:    "convert-to-pdf",
							Image:   image,
							Command: []string{"sh", "-c"},
							Args:    []string{fmt.Sprintf("pandoc -s -o /data/%s.pdf /data/text.md", pdfDoc.Spec.DocumentName)},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data-volume",
									MountPath: "/data",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:    "main",
							Image:   "alpine",
							Command: []string{"sh", "-c", "sleep 3600"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data-volume",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data-volume",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
	return j, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PdfDocumentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8sstartkubernetescomv2.PdfDocument{}).
		Named("pdfdocument").
		Complete(r)
}
