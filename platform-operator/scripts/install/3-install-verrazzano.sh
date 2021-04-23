#!/usr/bin/env bash
#
# Copyright (c) 2020, 2021, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#
SCRIPT_DIR=$(cd $(dirname "$0"); pwd -P)
. $SCRIPT_DIR/common.sh
. $SCRIPT_DIR/config.sh

CONFIG_DIR=$SCRIPT_DIR/config

VERRAZZANO_NS=verrazzano-system
VERRAZZANO_MC=verrazzano-mc
MONITORING_NS=monitoring

ENV_NAME=$(get_config_value ".environmentName")

INGRESS_TYPE=$(get_config_value ".ingress.type")
INGRESS_IP=$(get_verrazzano_ingress_ip)
if [ -n "${INGRESS_IP:-}" ]; then
  log "Found ingress address ${INGRESS_IP}"
else
  fail "Failed to find ingress address."
fi

DNS_TYPE=$(get_config_value ".dns.type")
DNS_SUFFIX=$(get_dns_suffix ${INGRESS_IP})

# Check if the nginx ingress ports are accessible
function check_ingress_ports() {
  exitvalue=0
  if [ ${INGRESS_TYPE} == "LoadBalancer" ] && [ $DNS_TYPE != "external" ]; then
    # Get the ports from the ingress
    PORTS=$(kubectl get services -n ingress-nginx ingress-controller-ingress-nginx-controller -o=custom-columns=PORT:.spec.ports[*].name --no-headers)
    IFS=',' read -r -a port_array <<< "$PORTS"

    index=0
    for element in "${port_array[@]}"
    do
      # For each get the port, nodePort and targetPort
      RESP=$(kubectl get services -n ingress-nginx ingress-controller-ingress-nginx-controller -o=custom-columns=PORT:.spec.ports[$index].port,NODEPORT:.spec.ports[$index].nodePort,TARGETPORT:.spec.ports[$index].targetPort --no-headers)
      ((index++))

      IFS=' ' read -r -a vals <<< "$RESP"
      PORT="${vals[0]}"
      NODEPORT="${vals[1]}"
      TARGETPORT="${vals[2]}"

      # Attempt to access the port on the $INGRESS_IP
      if [ $TARGETPORT == "https" ]; then
        ARGS=(-k https://$INGRESS_IP:$PORT)
        call_curl 0 response http_code ARGS
      else
        ARGS=(http://$INGRESS_IP:$PORT)
        call_curl 0 response http_code ARGS
      fi

      # Check the result of the curl call
      if [ $? -eq 0 ]; then
        log "Port $PORT is accessible on ingress address $INGRESS_IP.  Note that '404 page not found' is an expected response."
      else
        log "ERROR: Port $PORT is NOT accessible on ingress address $INGRESS_IP!  Check that security lists include an ingress rule for the node port $NODEPORT."
        log "See install README for details(https://github.com/verrazzano/verrazzano/operator/blob/master/install/README.md#1-oke-missing-security-list-ingress-rules)."
        exitvalue=1
      fi
    done
  fi
  return $exitvalue
}

action "Checking ingress ports" check_ingress_ports || fail "ERROR: Failed ingress port check."

set -eu

function install_verrazzano()
{
  if [ $(is_rancher_enabled) == "true" ]; then
    local RANCHER_HOSTNAME=rancher.${ENV_NAME}.${DNS_SUFFIX}

    local rancher_admin_password=`kubectl get secret --namespace cattle-system rancher-admin-secret -o jsonpath={.data.password} | base64 --decode`

    if [ -z "$rancher_admin_password" ] ; then
      error "ERROR: Failed to retrieve rancher-admin-secret - did you run the scripts to install Istio and system components?"
      return 1
    fi

    # Wait until rancher TLS cert is ready
    log "Waiting for Rancher TLS cert to reach ready state"
    kubectl wait --for=condition=ready cert tls-rancher-ingress -n cattle-system

    # Make sure rancher ingress has an IP
    wait_for_ingress_ip rancher cattle-system || exit 1

    get_rancher_access_token "${RANCHER_HOSTNAME}" "${rancher_admin_password}"
    if [ $? -ne 0 ] ; then
      error "ERROR: Failed to get rancher access token"
      exit 1
    fi
    local token_array=(${RANCHER_ACCESS_TOKEN//:/ })
  fi

  EXTRA_V8O_ARGUMENTS=$(get_verrazzano_helm_args_from_config)
  if [ ${REGISTRY_SECRET_EXISTS} == "TRUE" ]; then
    EXTRA_V8O_ARGUMENTS="${EXTRA_V8O_ARGUMENTS} --set global.imagePullSecrets[0]=${GLOBAL_IMAGE_PULL_SECRET}"
  fi

  local profile=$(get_install_profile)
  if [ ! -f "${VZ_CHARTS_DIR}/verrazzano/values.${profile}.yaml" ]; then
    error "The file ${VZ_CHARTS_DIR}/verrazzano/values.${profile}.yaml does not exist"
    exit 1
  fi
  local PROFILE_VALUES_OVERRIDE=" -f ${VZ_CHARTS_DIR}/verrazzano/values.${profile}.yaml"

  # Get the endpoint for the Kubernetes API server.  The endpoint returned has the format of IP:PORT
  local ENDPOINT=$(kubectl get endpoints --namespace default kubernetes --no-headers | awk '{ print $2}')
  local ENDPOINT_ARRAY=(${ENDPOINT//:/ })

  local DNS_TYPE=$(get_config_value ".dns.type")
  local EXTERNAL_DNS_ENABLED=false
  if [ "$DNS_TYPE" == "oci" ]; then
    EXTERNAL_DNS_ENABLED=true
  fi

  helm \
      upgrade --install verrazzano \
      ${VZ_CHARTS_DIR}/verrazzano \
      --namespace ${VERRAZZANO_NS} \
      --set image.pullPolicy=IfNotPresent \
      --set config.envName=${ENV_NAME} \
      --set config.dnsSuffix=${DNS_SUFFIX} \
      --set config.enableMonitoringStorage=true \
      --set kubernetes.service.endpoint.ip=${ENDPOINT_ARRAY[0]} \
      --set kubernetes.service.endpoint.port=${ENDPOINT_ARRAY[1]} \
      --set externaldns.enabled=${EXTERNAL_DNS_ENABLED} \
      --set keycloak.enabled=$(is_keycloak_enabled) \
      ${PROFILE_VALUES_OVERRIDE} \
      ${EXTRA_V8O_ARGUMENTS} || return $?

  log "Waiting for the verrazzano-operator pod in ${VERRAZZANO_NS} to reach Ready state"
  kubectl  wait -l app=verrazzano-operator --for=condition=Ready pod -n verrazzano-system

  log "Verifying that needed secrets are created"
  retries=0
  until [ "$retries" -ge 60 ]
  do
      kubectl get secret -n ${VERRAZZANO_NS} verrazzano | grep verrazzano && break
      retries=$(($retries+1))
      sleep 5
  done
  if ! kubectl get secret --namespace ${VERRAZZANO_NS} verrazzano ; then
      error "ERROR: failed creating verrazzano secret"
      exit 1
  fi
  log "Verrazzano install completed"
}

function install_oam_operator {

  log "Install OAM Kubernetes operator"
  helm upgrade --install --wait oam-kubernetes-runtime \
    ${CHARTS_DIR}/oam-kubernetes-runtime \
    --namespace "${VERRAZZANO_NS}" \
    -f $VZ_OVERRIDES_DIR/oam-kubernetes-runtime-values.yaml \
    || return $?
  if [ $? -ne 0 ]; then
    error "Failed to install OAM Kubernetes operator."
    return 1
  fi
}

function install_application_operator {

  log "Install Verrazzano Kubernetes application operator"
  helm upgrade --install --wait verrazzano-application-operator \
    $VZ_CHARTS_DIR/verrazzano-application-operator \
    --namespace "${VERRAZZANO_NS}" \
    -f $VZ_OVERRIDES_DIR/verrazzano-application-operator-values.yaml \
    ${EXTRA_V8O_ARGUMENTS} || return $?
  if [ $? -ne 0 ]; then
    error "Failed to install Verrazzano Kubernetes application operator."
    return 1
  fi
}

function install_weblogic_operator {

  log "Create WebLogic Kubernetes operator service account"
  kubectl create serviceaccount -n "${VERRAZZANO_NS}" weblogic-operator-sa
  if [ $? -ne 0 ]; then
    error "Failed to create WebLogic Kubernetes operator service account."
    return 1
  fi

  log "Install WebLogic Kubernetes operator"
  helm upgrade --install --wait weblogic-operator \
    ${CHARTS_DIR}/weblogic-operator \
    --namespace "${VERRAZZANO_NS}" \
    -f $VZ_OVERRIDES_DIR/weblogic-values.yaml \
    --set serviceAccount=weblogic-operator-sa \
    --set domainNamespaceSelectionStrategy=LabelSelector \
    --set domainNamespaceLabelSelector=verrazzano-managed \
    --set enableClusterRoleBinding=true \
    || return $?
  if [ $? -ne 0 ]; then
    error "Failed to install WebLogic Kubernetes operator."
    return 1
  fi
}

function install_coherence_operator {

  log "Install the Coherence Kubernetes operator"
  helm upgrade --install --wait coherence-operator \
    ${CHARTS_DIR}/coherence-operator \
    --namespace "${VERRAZZANO_NS}" \
    -f $VZ_OVERRIDES_DIR/coherence-values.yaml \
    || return $?
  if [ $? -ne 0 ]; then
    error "Failed to install the Coherence Kubernetes operator."
    return 1
  fi
}

# Set environment variable for checking if optional imagePullSecret was provided
REGISTRY_SECRET_EXISTS=$(check_registry_secret_exists)

if ! kubectl get namespace ${VERRAZZANO_NS} ; then
  action "Creating ${VERRAZZANO_NS} namespace" kubectl create namespace ${VERRAZZANO_NS} || exit 1
fi

log "Adding label needed by network policies to ${VERRAZZANO_NS} namespace"
kubectl label namespace ${VERRAZZANO_NS} "verrazzano.io/namespace=${VERRAZZANO_NS}" --overwrite

if ! kubectl get namespace ${VERRAZZANO_MC} ; then
  action "Creating ${VERRAZZANO_MC} namespace" kubectl create namespace ${VERRAZZANO_MC} || exit 1
fi

if ! kubectl get namespace ${MONITORING_NS} ; then
  action "Creating ${MONITORING_NS} namespace" kubectl create namespace ${MONITORING_NS} || exit 1
fi

log "Adding label needed by network policies to ${MONITORING_NS} namespace"
kubectl label namespace ${MONITORING_NS} "verrazzano.io/namespace=${MONITORING_NS}" --overwrite

# If Keycloak is being installed, create the Keycloak namespace if it doesn't exist so we can apply network policies
if [ $(is_keycloak_enabled) == "true" ] && ! kubectl get namespace keycloak ; then
  action "Creating keycloak namespace" kubectl create namespace keycloak || exit 1
fi

if [ "${REGISTRY_SECRET_EXISTS}" == "TRUE" ]; then
  if ! kubectl get secret ${GLOBAL_IMAGE_PULL_SECRET} -n ${VERRAZZANO_NS} > /dev/null 2>&1 ; then
    action "Copying ${GLOBAL_IMAGE_PULL_SECRET} secret to ${VERRAZZANO_NS} namespace" \
        copy_registry_secret "${VERRAZZANO_NS}"
  fi
fi

action "Installing Verrazzano system components" install_verrazzano || exit 1
action "Installing Coherence Kubernetes operator" install_coherence_operator || exit 1
action "Installing WebLogic Kubernetes operator" install_weblogic_operator || exit 1
action "Installing OAM Kubernetes operator" install_oam_operator || exit 1
action "Installing Verrazzano Application Kubernetes operator" install_application_operator || exit 1
