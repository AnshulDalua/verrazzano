// Copyright (c) 2022, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package weblogic

import (
	"flag"
	"testing"

	"github.com/onsi/ginkgo/v2"
)

var namespace string
var istioInjection string

func init() {
	flag.StringVar(&namespace, "namespace", generatedNamespace, "namespace is the app namespace")
	flag.StringVar(&istioInjection, "istioInjection", "enabled", "istioInjection enables the injection of istio side cars")
}

func TestWebLogicLogging(test *testing.T) {
	t.RegisterFailHandler()
	ginkgo.RunSpecs(test, "WebLogic Logging Suite")
}
