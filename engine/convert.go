// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import (
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func toPod(spec *Spec) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        spec.PodSpec.Name,
			Namespace:   spec.PodSpec.Namespace,
			Annotations: spec.PodSpec.Annotations,
			Labels:      spec.PodSpec.Labels,
		},
		Spec: v1.PodSpec{
			ServiceAccountName: spec.PodSpec.ServiceAccountName,
			RestartPolicy:      v1.RestartPolicyNever,
			Volumes:            toVolumes(spec),
			Containers:         toContainers(spec),
			NodeName:           spec.PodSpec.NodeName,
			NodeSelector:       spec.PodSpec.NodeSelector,
			Tolerations:        toTolerations(spec),
			ImagePullSecrets:   toImagePullSecrets(spec),
			HostAliases:        toHostAliases(spec),
			DNSConfig:          toDnsConfig(spec),
		},
	}
}

func toDnsConfig(spec *Spec) *v1.PodDNSConfig {
	var dnsOptions []v1.PodDNSConfigOption
	if len(spec.PodSpec.DnsConfig.Options) > 0 {
		for _, option := range spec.PodSpec.DnsConfig.Options {
			o := v1.PodDNSConfigOption{
				Name:  option.Name,
				Value: option.Value,
			}
			dnsOptions = append(dnsOptions, o)
		}
	}
	return &v1.PodDNSConfig{
		Nameservers: spec.PodSpec.DnsConfig.Nameservers,
		Searches:    spec.PodSpec.DnsConfig.Searches,
		Options:     dnsOptions,
	}
}

func toHostAliases(spec *Spec) []v1.HostAlias {
	var hostAliases []v1.HostAlias
	for _, hostAlias := range spec.PodSpec.HostAliases {
		if len(hostAlias.Hostnames) > 0 {
			hostAliases = append(hostAliases, v1.HostAlias{
				IP:        hostAlias.IP,
				Hostnames: hostAlias.Hostnames,
			})
		}
	}
	return hostAliases
}

func toTolerations(spec *Spec) []v1.Toleration {
	var tolerations []v1.Toleration
	for _, toleration := range spec.PodSpec.Tolerations {
		t := v1.Toleration{
			Key:      toleration.Key,
			Operator: v1.TolerationOperator(toleration.Operator),
			Effect:   v1.TaintEffect(toleration.Effect),
			Value:    toleration.Value,
		}
		if toleration.TolerationSeconds != nil {
			t.TolerationSeconds = int64ptr(int64(*toleration.TolerationSeconds))
		}
		tolerations = append(tolerations, t)
	}
	return tolerations
}

func toVolumes(spec *Spec) []v1.Volume {
	var volumes []v1.Volume
	for _, v := range spec.Volumes {
		if v.EmptyDir != nil {
			source := &v1.EmptyDirVolumeSource{}
			if strings.EqualFold(v.EmptyDir.Medium, "memory") {
				source.Medium = v1.StorageMediumMemory
				if v.EmptyDir.SizeLimit > int64(0) {
					source.SizeLimit = resource.NewQuantity(v.EmptyDir.SizeLimit, resource.BinarySI)
				}
			}
			volume := v1.Volume{
				Name: v.EmptyDir.ID,
				VolumeSource: v1.VolumeSource{
					EmptyDir: source,
				},
			}
			volumes = append(volumes, volume)
		}

		if v.HostPath != nil {
			hostPathType := v1.HostPathDirectoryOrCreate
			volume := v1.Volume{
				Name: v.HostPath.ID,
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: v.HostPath.Path,
						Type: &hostPathType,
					},
				},
			}
			volumes = append(volumes, volume)
		}

		if v.Claim != nil {
			volume := v1.Volume{
				Name: v.Claim.ID,
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: v.Claim.ClaimName,
						ReadOnly:  true,
					},
				},
			}
			volumes = append(volumes, volume)
		}

		if v.DownwardAPI != nil {
			var items []v1.DownwardAPIVolumeFile

			for _, item := range v.DownwardAPI.Items {
				items = append(items, v1.DownwardAPIVolumeFile{
					Path: item.Path,
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: item.FieldPath,
					},
				})
			}

			volume := v1.Volume{
				Name: v.DownwardAPI.ID,
				VolumeSource: v1.VolumeSource{
					DownwardAPI: &v1.DownwardAPIVolumeSource{
						Items: items,
					},
				},
			}

			volumes = append(volumes, volume)
		}
	}

	return volumes
}

func toContainers(spec *Spec) []v1.Container {
	var containers []v1.Container

	for _, s := range spec.Steps {
		container := v1.Container{
			Name:            s.ID,
			Image:           s.Placeholder,
			Command:         s.Entrypoint,
			Args:            s.Command,
			ImagePullPolicy: toPullPolicy(s.Pull),
			WorkingDir:      s.WorkingDir,
			Resources:       toResources(s.Resources),
			SecurityContext: &v1.SecurityContext{
				Privileged: boolptr(s.Privileged),
			},
			VolumeMounts: toVolumeMounts(spec, s),
			Env:          toEnv(spec, s),
		}

		containers = append(containers, container)
	}

	return containers
}

func toEnv(spec *Spec, step *Step) []v1.EnvVar {
	var envVars []v1.EnvVar

	for k, v := range step.Envs {
		envVars = append(envVars, v1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	for _, secret := range step.Secrets {
		envVars = append(envVars, v1.EnvVar{
			Name: secret.Env,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: spec.PodSpec.Name,
					},
					Key:      secret.Name,
					Optional: boolptr(true),
				},
			},
		})
	}

	envVars = append(envVars, v1.EnvVar{
		Name: "KUBERNETES_NODE",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "spec.nodeName",
			},
		},
	})

	return envVars
}

func toEnvFrom(step *Step) []v1.EnvFromSource {
	return []v1.EnvFromSource{
		{
			SecretRef: &v1.SecretEnvSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: step.ID,
				},
			},
		},
	}
}

func toSecret(spec *Spec) *v1.Secret {
	stringData := make(map[string]string)
	for _, secret := range spec.Secrets {
		stringData[secret.Name] = secret.Data
	}

	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: spec.PodSpec.Name,
		},
		Type:       "Opaque",
		StringData: stringData,
	}
}

func toDockerConfigSecret(spec *Spec) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: spec.PullSecret.Name,
		},
		Type: "kubernetes.io/dockerconfigjson",
		StringData: map[string]string{
			".dockerconfigjson": spec.PullSecret.Data,
		},
	}
}

func toImagePullSecrets(spec *Spec) []v1.LocalObjectReference {
	var pullSecrets []v1.LocalObjectReference
	if spec.PullSecret != nil {
		pullSecrets = []v1.LocalObjectReference{{
			Name: spec.PullSecret.Name,
		}}
	}
	return pullSecrets
}

func toVolumeMounts(spec *Spec, step *Step) []v1.VolumeMount {
	var volumeMounts []v1.VolumeMount
	for _, v := range step.Volumes {
		id, ok := lookupVolumeID(spec, v.Name)
		if !ok {
			continue
		}
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      id,
			MountPath: v.Path,
			ReadOnly:  v.ReadOnly,
		})
	}

	return volumeMounts
}

func toResources(src Resources) v1.ResourceRequirements {
	var dst v1.ResourceRequirements
	if src.Limits.Memory > 0 || src.Limits.CPU > 0 {
		dst.Limits = v1.ResourceList{}
		if src.Limits.Memory > int64(0) {
			dst.Limits[v1.ResourceMemory] = *resource.NewQuantity(
				src.Limits.Memory, resource.BinarySI)
		}
		if src.Limits.CPU > int64(0) {
			dst.Limits[v1.ResourceCPU] = *resource.NewMilliQuantity(
				src.Limits.CPU, resource.DecimalSI)
		}
	}
	if src.Requests.Memory > 0 || src.Requests.CPU > 0 {
		dst.Requests = v1.ResourceList{}
		if src.Requests.Memory > int64(0) {
			dst.Requests[v1.ResourceMemory] = *resource.NewQuantity(
				src.Requests.Memory, resource.BinarySI)
		}
		if src.Requests.CPU > int64(0) {
			dst.Requests[v1.ResourceCPU] = *resource.NewMilliQuantity(
				src.Requests.CPU, resource.DecimalSI)
		}
	}
	return dst
}

// LookupVolume is a helper function that will lookup
// the id for a volume.
func lookupVolumeID(spec *Spec, name string) (string, bool) {
	for _, v := range spec.Volumes {
		if v.EmptyDir != nil && v.EmptyDir.Name == name {
			return v.EmptyDir.ID, true
		}

		if v.HostPath != nil && v.HostPath.Name == name {
			return v.HostPath.ID, true
		}

		if v.Claim != nil && v.Claim.Name == name {
			return v.Claim.ID, true
		}

		if v.DownwardAPI != nil && v.DownwardAPI.Name == name {
			return v.DownwardAPI.ID, true
		}
	}

	return "", false
}

func toPullPolicy(policy PullPolicy) v1.PullPolicy {
	switch policy {
	case PullAlways:
		return v1.PullAlways
	case PullNever:
		return v1.PullNever
	case PullIfNotExists:
		return v1.PullIfNotPresent
	default:
		return v1.PullIfNotPresent
	}
}

func int64ptr(v int64) *int64 {
	return &v
}

func boolptr(v bool) *bool {
	return &v
}

func stringptr(v string) *string {
	return &v
}
