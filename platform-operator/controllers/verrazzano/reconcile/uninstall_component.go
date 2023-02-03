// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package reconcile

import (
	ctrlerrors "github.com/verrazzano/verrazzano/pkg/controller/errors"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1alpha1"
	vzconst "github.com/verrazzano/verrazzano/platform-operator/constants"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/rancher"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/registry"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// compStateUpgradeStart is the state when a component is starting the Uninstall flow
	compStateUninstallStart componentState = "compStateUninstallStart"

	// compStatePreUninstall is the state when a component does a pre-Uninstall
	compStatePreUninstall componentState = "compStatePreUninstall"

	// compStateUninstall is the state where a component does an Uninstall
	compStateUninstall componentState = "compStateUninstall"

	// compStateWaitUninstalled is the state when a component is waiting to be uninstalled
	compStateWaitUninstalled componentState = "compStateWaitUninstalled"

	// compStateUninstalleDone is the state when component Uninstall is done
	compStateUninstalleDone componentState = "compStateUninstalleDone"

	// compStateUninstallEnd is the terminal state
	compStateUninstallEnd componentState = "compStateUninstallEnd"
)

// componentUninstallContext has the Uninstall context for a Verrazzano component Uninstall
type componentUninstallContext struct {
	state componentState
}

// UninstallComponents will Uninstall the components as required
func (r *Reconciler) uninstallComponents(log vzlog.VerrazzanoLogger, cr *v1alpha1.Verrazzano, tracker *UninstallTracker) (ctrl.Result, error) {
	spiCtx, err := spi.NewContext(log, r.Client, cr, nil, r.DryRun)
	if err != nil {
		return newRequeueWithDelay(), err
	}

	var requeue bool

	// Loop through the Verrazzano components and Uninstall each one.
	// Don't block uninstalling the next component if the current one has an error.
	// It is normal for a component to return an error if it is waiting for some condition.
	for _, comp := range registry.GetComponents() {
		UninstallContext := tracker.getComponentUninstallContext(comp.Name())
		result, err := r.uninstallSingleComponent(spiCtx, UninstallContext, comp)
		if err != nil || result.Requeue {
			requeue = true
		}

	}
	if requeue {
		return newRequeueWithDelay(), nil
	}

	// All components have been Uninstalled
	return ctrl.Result{}, nil
}

// UninstallSingleComponent Uninstalls a single component
func (r *Reconciler) uninstallSingleComponent(spiCtx spi.ComponentContext, UninstallContext *componentUninstallContext, comp spi.Component) (ctrl.Result, error) {
	compName := comp.Name()
	compContext := spiCtx.Init(compName).Operation(vzconst.UninstallOperation)

	for UninstallContext.state != compStateUninstallEnd {
		var err error
		UninstallContext.state, err = r.executeComponentUninstallState(compContext, comp, UninstallContext.state)
		if err != nil {
			return newRequeueWithDelay(), err
		}
	}
	// Component has been Uninstalled
	return ctrl.Result{}, nil
}

// executeComponentUninstallState Manages the uninstall state machine for a component
func (r *Reconciler) executeComponentUninstallState(compContext spi.ComponentContext, comp spi.Component, currentState componentState) (componentState, error) {
	compName := comp.Name()
	compLog := compContext.Log()
	rancherProvisioned, err := rancher.IsClusterProvisionedByRancher()
	if err != nil {
		return "", err
	}

	var nextState componentState
	switch currentState {
	case compStateUninstallStart:
		// Check if operator based uninstall is supported
		if !comp.IsOperatorUninstallSupported() {
			return compStateUninstallEnd, nil
		}
		if comp.Name() == rancher.ComponentName && rancherProvisioned {
			compLog.Oncef("Cluster was provisioned by Rancher. Component %s will not be uninstalled.", rancher.ComponentName)
			return compStateUninstallEnd, nil
		}
		// Check if component is installed, if not continue
		installed, err := comp.IsInstalled(compContext)
		if err != nil {
			compLog.Errorf("Failed checking if component %s is installed: %v", compName, err)
			return compStateUninstallEnd, nil
		}
		if !installed {
			compLog.Oncef("Component %s is not installed, nothing to do for uninstall", compName)
			return compStateUninstallEnd, nil
		}
		if err := r.updateComponentStatus(compContext, "Uninstall started", v1alpha1.CondUninstallStarted); err != nil {
			return "", err
		}
		compLog.Oncef("Component %s is starting to uninstall", compName)
		nextState = compStatePreUninstall

	case compStatePreUninstall:
		compLog.Oncef("Component %s is calling pre-uninstall", compName)
		if err := comp.PreUninstall(compContext); err != nil {
			// Components will log errors, could be waiting for condition
			return currentState, err
		}
		nextState = compStateUninstall

	case compStateUninstall:
		compLog.Progressf("Component %s is calling uninstall", compName)
		if err := comp.Uninstall(compContext); err != nil {
			if !ctrlerrors.IsRetryableError(err) {
				compLog.Errorf("Failed uninstalling component %s, will retry: %v", compName, err)
			}
			return "", err
		}
		nextState = compStateWaitUninstalled

	case compStateWaitUninstalled:
		installed, err := comp.IsInstalled(compContext)
		if err != nil {
			compLog.Errorf("Failed checking if component %s is installed: %v", compName, err)
			return currentState, err
		}
		if installed {
			compLog.Progressf("Waiting for component %s to be uninstalled", compName)
			return currentState, nil
		}
		compLog.Progressf("Component %s has been uninstalled, running post-uninstall", compName)
		if err := comp.PostUninstall(compContext); err != nil {
			if !ctrlerrors.IsRetryableError(err) {
				compLog.Errorf("PostUninstall for component %s failed: %v", compName, err)
			}
			return currentState, nil
		}
		nextState = compStateUninstalleDone

	case compStateUninstalleDone:
		if err := r.updateComponentStatus(compContext, "Uninstall complete", v1alpha1.CondUninstallComplete); err != nil {
			return currentState, err
		}
		compLog.Oncef("Component %s has successfully uninstalled", compName)
		nextState = compStateUninstallEnd
	}
	return nextState, nil
}

// getComponentUninstallContext gets the Uninstall context for the component
func (vuc *UninstallTracker) getComponentUninstallContext(compName string) *componentUninstallContext {
	context, ok := vuc.compMap[compName]
	if !ok {
		context = &componentUninstallContext{
			state: compStateUninstallStart,
		}
		vuc.compMap[compName] = context
	}
	return context
}
