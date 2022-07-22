package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"coralogixClient/clients"
	v1 "coralogixClient/clients/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime2 "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// Operator for checking events
type Operator struct {
	prometheusMiddleware *clients.PrometheusMiddleware
	coralogixClient      clients.CoralogixClient
	podsInformer         cache.Controller
	logsStore            cache.Store
	logsInformer         cache.Controller
}

type PodLog struct {
	Pod       corev1.Pod
	EventType watch.EventType
}

func NewCoralogixOperator() (*Operator, error) {
	coralogixClient := clients.CoralogixClient{}
	prometheusMiddleware := clients.NewPrometheusMiddleware()
	op := Operator{
		coralogixClient:      coralogixClient,
		prometheusMiddleware: prometheusMiddleware,
	}

	log.Print("Shared Informer app started")

	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.Panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panic(err.Error())
	}

	factory := informers.NewSharedInformerFactory(clientset, 10*time.Second)
	podsInformer := factory.Core().V1().Pods().Informer()
	podsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.AddPod,
		DeleteFunc: op.DeletePod,
	})
	op.podsInformer = podsInformer

	v1Clientset, err := v1.NewForConfig(config)
	if err != nil {
		log.Panic(err.Error())
	}

	v1.AddToScheme(scheme.Scheme)

	op.logsStore, op.logsInformer = cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime2.Object, err error) {
				return v1Clientset.Logs("").List(context.TODO(), lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return v1Clientset.Logs("").Watch(context.TODO(), lo)
			},
		},
		&v1.Log{},
		1*time.Minute,
		cache.ResourceEventHandlerFuncs{
			AddFunc: op.AddLog,
		},
	)

	return &op, nil
}

// AddPod function to add a new object to the queue in case of creating a resource
func (op *Operator) AddPod(obj interface{}) {
	(*op.prometheusMiddleware.Gauge).Inc()
	op.OnPodChange(obj, watch.Added)
}

// DeletePod function to add an object to the queue in case of deleting a resource
func (op *Operator) DeletePod(old interface{}) {
	(*op.prometheusMiddleware.Gauge).Dec()
	op.OnPodChange(old, watch.Deleted)
}

// AddLog function to add a new object to the queue in case of creating a resource
func (op *Operator) AddLog(obj interface{}) {
	(*op.prometheusMiddleware.Gauge).Inc()
	logJson, err := json.Marshal(obj.(*v1.Log))
	if err != nil {
		fmt.Println(err)
		return
	}
	op.coralogixClient.SendMsg(string(logJson))
}

func (op *Operator) OnPodChange(obj interface{}, eventType watch.EventType) {
	pod := obj.(*corev1.Pod)
	podLog := &PodLog{
		Pod:       *pod,
		EventType: eventType,
	}
	podLogJson, err := json.Marshal(podLog)
	if err != nil {
		fmt.Println(err)
		return
	}
	op.coralogixClient.SendMsg(string(podLogJson))
}

func (op *Operator) Run() {
	go op.prometheusMiddleware.Run()

	stopper := make(chan struct{})
	defer close(stopper)
	defer runtime.HandleCrash()

	go op.podsInformer.Run(stopper)
	if !cache.WaitForCacheSync(stopper, op.podsInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	go op.logsInformer.Run(stopper)
	if !cache.WaitForCacheSync(stopper, op.logsInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}
	<-stopper
}
