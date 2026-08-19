package barbican

import (
	barbicanv1beta1 "github.com/openstack-k8s-operators/barbican-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/lib-common/modules/common/pod"
	"github.com/openstack-k8s-operators/lib-common/modules/users"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// DbSyncJob func
func DbSyncJob(instance *barbicanv1beta1.Barbican, labels map[string]string, annotations map[string]string) *batchv1.Job {
	// The dbsync job just needs the main barbican config files
	dbSyncVolumes, dbSyncMounts := GetOsloConfigVolumes(instance.Name, DBSyncConfigVolume)

	// add CA cert if defined
	if instance.Spec.BarbicanAPI.TLS.CaBundleSecretName != "" {
		dbSyncVolumes = append(dbSyncVolumes, instance.Spec.BarbicanAPI.TLS.CreateVolume())
		dbSyncMounts = append(dbSyncMounts, instance.Spec.BarbicanAPI.TLS.CreateVolumeMounts(nil)...)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-db-sync",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					AutomountServiceAccountToken: ptr.To(false),
					RestartPolicy:                corev1.RestartPolicyOnFailure,
					ServiceAccountName:           instance.RbacResourceName(),
					SecurityContext:              pod.RestrictivePodSecurityContext(users.BarbicanUID, users.BarbicanGID),
					Volumes:                      dbSyncVolumes,
					Containers: []corev1.Container{
						{
							Name: instance.Name + "-db-sync",
							Command: []string{
								"barbican-manage",
							},
							Args:            []string{"db", "upgrade"},
							Image:           instance.Spec.BarbicanAPI.ContainerImage,
							SecurityContext: pod.RestrictiveSecurityContext(users.BarbicanUID, users.BarbicanGID),
							Env:             []corev1.EnvVar{},
							VolumeMounts:    dbSyncMounts,
						},
					},
				},
			},
		},
	}

	if instance.Spec.NodeSelector != nil {
		job.Spec.Template.Spec.NodeSelector = *instance.Spec.NodeSelector
	}

	return job
}
