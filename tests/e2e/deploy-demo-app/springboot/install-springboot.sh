#!/usr/bin/env bash
#
# Copyright (c) 2020, 2021, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#
SCRIPT_DIR=$(cd $(dirname $0); pwd -P)

set -u

NAMESPACE="oam-springboot"
SPRINGBOOT_COMPONENT_FILE="${SCRIPT_DIR}/springboot-comp.yaml"

echo "Installing Springboot OAM application."

status=$(kubectl get namespace ${NAMESPACE} -o jsonpath="{.status.phase}" 2> /dev/null)
if [ "${status}" == "Active" ]; then
  echo "Found namespace ${NAMESPACE}."
else
  echo "Create namespace ${NAMESPACE}."
  kubectl create namespace "${NAMESPACE}"
  if [ $? -ne 0 ]; then
      echo "ERROR: Failed to create namespace ${NAMESPACE}, exiting."
      exit 1
  fi
fi

# QUESTION : Is it a good option to copy .yaml files to /tmp/springboot before replacing the tokens
# and call kubectl apply -f /tmp/springboot/ ?
echo "Substitute image name template in ${SPRINGBOOT_COMPONENT_FILE} as ${SPRINGBOOT_APP_IMAGE}"
sed -i '' -e "s|%SPRINGBOOT_APP_IMAGE%|${SPRINGBOOT_APP_IMAGE}|" ${SPRINGBOOT_COMPONENT_FILE}

echo "Apply application configuration."
kubectl apply -f ${SCRIPT_DIR}/
code=$?
if [ ${code} -ne 0 ]; then
  echo "ERROR: Applying application configuration failed: ${code}. Exiting."
  exit ${code}
fi

echo "Wait for at least one running workload pod."
attempt=1
while true; do
  kubectl -n "${NAMESPACE}" wait --for=condition=ready pods --selector='app.oam.dev/name=springboot-appconf' --timeout 15s
  if [ $? -eq 0 ]; then
    echo "Application pods found ready on attempt ${attempt}."
    break
  elif [ ${attempt} -eq 1 ]; then
    echo "No application pods found ready on initial attempt. Retrying after delay."
  elif [ ${attempt} -ge 30 ]; then
    echo "ERROR: No application pod found ready after ${attempt} attempts. Listing pods."
    kubectl get pods -n "${NAMESPACE}"
    echo "ERROR: Exiting."
    exit 1
  fi
  attempt=$(($attempt+1))
  sleep .5
done

echo "Installation of Springboot OAM application successful."
