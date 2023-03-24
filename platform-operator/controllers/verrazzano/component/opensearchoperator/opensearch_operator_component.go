// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package opensearchoperator

import (
	"context"
	"github.com/verrazzano/verrazzano/pkg/k8s/ready"
	"github.com/verrazzano/verrazzano/pkg/vzcr"
	"github.com/verrazzano/verrazzano/platform-operator/constants"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/certmanager"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/helm"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	"github.com/verrazzano/verrazzano/platform-operator/internal/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// ComponentName is the name of the component
	ComponentName = "opensearch-operator"

	// ComponentNamespace is the namespace of the component
	ComponentNamespace = constants.VerrazzanoLoggingNamespace

	// ComponentJSONName is the json name of the opensearch-operator component in CRD
	ComponentJSONName = "opensearchOperator"
)

type opensearchOperatorComponent struct {
	helm.HelmComponent
}

func NewComponent() spi.Component {
	return opensearchOperatorComponent{
		HelmComponent: helm.HelmComponent{
			ReleaseName:               ComponentName,
			JSONName:                  ComponentJSONName,
			ChartDir:                  filepath.Join(config.GetThirdPartyDir(), ComponentName),
			ChartNamespace:            ComponentNamespace,
			ValuesFile:                filepath.Join(config.GetHelmOverridesDir(), "opensearch-operator-values.yaml"),
			IgnoreNamespaceOverride:   true,
			SupportsOperatorInstall:   true,
			SupportsOperatorUninstall: true,
			Dependencies:              []string{certmanager.ComponentName},
			AvailabilityObjects: &ready.AvailabilityObjects{
				DeploymentNames: getDeploymentList(),
			},
			IngressNames:            getIngressList(),
			GetInstallOverridesFunc: GetOverrides,
			//AppendOverridesFunc:     AppendOverrides,
		},
	}
}

// IsEnabled returns true if the component is enabled for install
func (o opensearchOperatorComponent) IsEnabled(effectiveCr runtime.Object) bool {
	return vzcr.IsOpenSearchOperatorEnabled(effectiveCr)
}

// IsReady - component specific ready-check
func (o opensearchOperatorComponent) IsReady(context spi.ComponentContext) bool {
	if o.HelmComponent.IsReady(context) {
		return o.isReady(context)
	}
	return false
}

// PreInstall runs before components are installed
func (o opensearchOperatorComponent) PreInstall(ctx spi.ComponentContext) error {
	cli := ctx.Client()
	log := ctx.Log()

	// create namespace
	ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ComponentNamespace}}
	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}

	ns.Labels["verrazzano.io/namespace"] = ComponentNamespace
	if _, err := controllerutil.CreateOrUpdate(context.TODO(), cli, &ns, func() error {
		return nil
	}); err != nil {
		return log.ErrorfNewErr("Failed to create or update the %s namespace: %v", ComponentNamespace, err)
	}

	err := createSecurityconfigSecret(ctx)
	if err != nil {
		return err
	}

	return o.HelmComponent.PreInstall(ctx)
}

// Install OpenSearchOperator install processing
func (o opensearchOperatorComponent) Install(ctx spi.ComponentContext) error {
	if err := o.HelmComponent.Install(ctx); err != nil {
		return err
	}
	return nil
}

//func (o opensearchOperatorComponent) Reconcile(ctx spi.ComponentContext) error {
//	if err := createSecurityconfigSecret(ctx); err != nil {
//		return err
//	}
//	return nil
//}

func (o opensearchOperatorComponent) PostInstall(ctx spi.ComponentContext) error {
	//if err := checkServiceExists(ctx, opsterOSService); err != nil {
	//	return err
	//}

	if err := createIngress(ctx, opsterOSIngressName); err != nil {
		return err
	}

	//if err := checkServiceExists(ctx, opsterOSDService); err != nil {
	//	return err
	//}

	if err := createIngress(ctx, opsterOSDIngressName); err != nil {
		return err
	}

	return nil
}

// MonitorOverrides checks whether monitoring of install overrides is enabled or not
func (c opensearchOperatorComponent) MonitorOverrides(ctx spi.ComponentContext) bool {
	if ctx.EffectiveCR().Spec.Components.OpenSearchOperator != nil {
		if ctx.EffectiveCR().Spec.Components.OpenSearchOperator.MonitorChanges != nil {
			return *ctx.EffectiveCR().Spec.Components.OpenSearchOperator.MonitorChanges
		}
		return true
	}
	return false
}
