// Copyright (c) 2020, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"net/http"

	v1alpha1 "github.com/verrazzano/verrazzano/platform-operator/apis/platform/v1alpha1"
	"github.com/verrazzano/verrazzano/platform-operator/clientset/versioned/scheme"
	rest "k8s.io/client-go/rest"
)

type PlatformV1alpha1Interface interface {
	RESTClient() rest.Interface
	ModuleDefinitionsGetter
	OperatorDefinitionsGetter
	PlatformsGetter
}

// PlatformV1alpha1Client is used to interact with features provided by the platform group.
type PlatformV1alpha1Client struct {
	restClient rest.Interface
}

func (c *PlatformV1alpha1Client) ModuleDefinitions(namespace string) ModuleDefinitionInterface {
	return newModuleDefinitions(c, namespace)
}

func (c *PlatformV1alpha1Client) OperatorDefinitions(namespace string) OperatorDefinitionInterface {
	return newOperatorDefinitions(c, namespace)
}

func (c *PlatformV1alpha1Client) Platforms(namespace string) PlatformInterface {
	return newPlatforms(c, namespace)
}

// NewForConfig creates a new PlatformV1alpha1Client for the given config.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*PlatformV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	httpClient, err := rest.HTTPClientFor(&config)
	if err != nil {
		return nil, err
	}
	return NewForConfigAndClient(&config, httpClient)
}

// NewForConfigAndClient creates a new PlatformV1alpha1Client for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*PlatformV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &PlatformV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new PlatformV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *PlatformV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new PlatformV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *PlatformV1alpha1Client {
	return &PlatformV1alpha1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *PlatformV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
