package v1

import (
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type V1Interface interface {
	Logs(namespace string) LogInterface
}

type V1Client struct {
	restClient rest.Interface
}

func NewForConfig(c *rest.Config) (*V1Client, error) {
	config := *c
	config.ContentConfig.GroupVersion = &SchemeGroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &V1Client{restClient: client}, nil
}

func (c *V1Client) Logs(namespace string) LogInterface {
	return &logClient{
		restClient: c.restClient,
		ns:         namespace,
	}
}
