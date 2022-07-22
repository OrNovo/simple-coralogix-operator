package v1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type LogInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*LogList, error)
	Get(ctx context.Context, name string, options metav1.GetOptions) (*Log, error)
	Create(ctx context.Context, log *Log) (*Log, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

type logClient struct {
	restClient rest.Interface
	ns         string
}

func (c *logClient) List(ctx context.Context, opts metav1.ListOptions) (*LogList, error) {
	result := LogList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("logs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *logClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*Log, error) {
	result := Log{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("logs").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *logClient) Create(ctx context.Context, log *Log) (*Log, error) {
	result := Log{}
	err := c.restClient.
		Post().
		Namespace(c.ns).
		Resource("logs").
		Body(log).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *logClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource("logs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
}
