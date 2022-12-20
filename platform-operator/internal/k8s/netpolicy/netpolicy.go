// Copyright (c) 2021, 2022, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package netpolicy

import (
	"context"
	"github.com/verrazzano/verrazzano/platform-operator/constants"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	networkPolicyAPIVersion  = "networking.k8s.io/v1"
	networkPolicyKind        = "NetworkPolicy"
	networkPolicyPodName     = "verrazzano-platform-operator"
	podAppLabel              = "app"
	verrazzanoNamespaceLabel = "verrazzano.io/namespace"
	appNameLabel             = "app.kubernetes.io/name"
)

// CreateOrUpdateNetworkPolicies creates or updates network policies for the platform operator to
// limit network ingress.
func CreateOrUpdateNetworkPolicies(client client.Client) (controllerutil.OperationResult, error) {
	netPolicy := newNetworkPolicy()
	objKey := &netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: netPolicy.ObjectMeta.Name, Namespace: netPolicy.ObjectMeta.Namespace}}

	opResult, err := controllerutil.CreateOrUpdate(context.TODO(), client, objKey, func() error {
		netPolicy.Spec.DeepCopyInto(&objKey.Spec)
		return nil
	})

	return opResult, err
}

// newNetworkPolicy returns a populated NetworkPolicy with ingress rules for this operator.
func newNetworkPolicy() *netv1.NetworkPolicy {
	tcpProtocol := corev1.ProtocolTCP
	webhookPort := intstr.FromInt(9443)
	metricsPort := intstr.FromInt(9100)

	return &netv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: networkPolicyAPIVersion,
			Kind:       networkPolicyKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.VerrazzanoInstallNamespace,
			Name:      networkPolicyPodName,
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					podAppLabel: networkPolicyPodName,
				},
			},
			PolicyTypes: []netv1.PolicyType{
				netv1.PolicyTypeIngress,
			},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					// ingress from kubernetes API server for webhooks
					Ports: []netv1.NetworkPolicyPort{
						{
							Protocol: &tcpProtocol,
							Port:     &webhookPort,
						},
					},
				},
				{
					From: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									verrazzanoNamespaceLabel: constants.VerrazzanoMonitoringNamespace,
								},
							},
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									appNameLabel: constants.PrometheusStorageLabelValue,
								},
							},
						},
					},
					// ingress from Prometheus server for scraping metrics
					Ports: []netv1.NetworkPolicyPort{
						{
							Protocol: &tcpProtocol,
							Port:     &metricsPort,
						},
					},
				},
			}},
	}
}
