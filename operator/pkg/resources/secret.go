//
// Copyright 2020 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package resources

import (
	researchv1alpha1 "github.com/IBM/integrity-enforcer/operator/pkg/apis/research/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//regkey.yaml
func BuildRegKeySecretForCR(cr *researchv1alpha1.IntegrityEnforcer) *corev1.Secret {
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.RegKeySecret.Name,
			Namespace: cr.Namespace,
		},
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: cr.Spec.RegKeySecret.Value,
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}
	return sec
}

// //server-secret.yaml
func BuildKeyringSecretForIEFromValue(cr *researchv1alpha1.IntegrityEnforcer) *corev1.Secret {
	metaLabels := map[string]string{
		"app":                    cr.Name,
		"app.kubernetes.io/name": cr.Spec.CertPool.Name,
		// "app.kubernetes.io/component":  instance.ReleaseName(),
		"app.kubernetes.io/managed-by": "operator",
		// "app.kubernetes.io/instance":   instance.ReleaseName(),
		// "release":                      instance.ReleaseName(),
		"role": "security",
	}
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.CertPool.Name,
			Namespace: cr.Namespace,
			Labels:    metaLabels,
		},
		Data: make(map[string][]byte),
		Type: corev1.SecretTypeOpaque,
	}
	return sec
}

// ie-server-tls
func BuildTlsSecretForIE(cr *researchv1alpha1.IntegrityEnforcer) *corev1.Secret {
	var empty []byte
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.WebhookServerTlsSecretName,
			Namespace: cr.Namespace,
		},
		Data: map[string][]byte{
			corev1.TLSCertKey:       empty, // "tls.crt"
			corev1.TLSPrivateKeyKey: empty,
			"ca.crt":                empty,
		},
		Type: corev1.SecretTypeTLS,
	}
	return sec
}
