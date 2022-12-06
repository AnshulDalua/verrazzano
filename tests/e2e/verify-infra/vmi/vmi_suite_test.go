// Copyright (c) 2020, 2022, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package vmi_test

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
)

func TestVmi(test *testing.T) {
	t.RegisterFailHandler()
	ginkgo.RunSpecs(test, "VMI Suite")
}
