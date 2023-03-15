#!/bin/bash
# Copyright (c) 2020, 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#While loop for verrazzano-platform-operator to wait for webhooks to be started before starting up
SCRIPT_DIR=$(
    cd $(dirname "$0")
    pwd -P
)
${SCRIPT_DIR}/poll_webhook.sh "https://verrazzano-platform-operator-webhook:443/validate-install-verrazzano-io-v1alpha1-verrazzano"
${SCRIPT_DIR}/poll_webhook.sh "https://verrazzano-platform-operator-webhook:443/validate-install-verrazzano-io-v1beta1-verrazzano"
${SCRIPT_DIR}/poll_webhook.sh "https://verrazzano-platform-operator-webhook:443/v1beta1-validate-mysql-install-override-values"
${SCRIPT_DIR}/poll_webhook.sh "https://verrazzano-platform-operator-webhook:443/v1alpha1-validate-mysql-install-override-values"
${SCRIPT_DIR}/poll_webhook.sh "https://verrazzano-platform-operator-webhook:443/v1beta1-validate-requirements"
${SCRIPT_DIR}/poll_webhook.sh "https://verrazzano-platform-operator-webhook:443/v1alpha1-validate-requirements"
${SCRIPT_DIR}/poll_webhook.sh "-XPOST https://verrazzano-platform-operator-webhook:443/convert -d '{\"apiVersion\":\"apiextensions.k8s.io/v1\", \"kind\":\"ConversionReview\", \"request\":{}}')"
