package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"coralogixClient/clients"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// Operator for checking events
type Operator struct {
	prometheusMiddleware *clients.PrometheusMiddleware
	coralogixClient      clients.CoralogixClient
	informer             cache.Controller
}

type PodLog struct {
	Pod       corev1.Pod
	EventType watch.EventType
}

func NewCoralogixOperator() (*Operator, error) {
	coralogixClient := clients.CoralogixClient{}

	prometheusMiddleware := clients.NewPrometheusMiddleware()
	go prometheusMiddleware.Run()

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
	informer := factory.Core().V1().Pods().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    op.Add,
		DeleteFunc: op.Delete,
	})
	op.informer = informer

	return &op, nil
}

// Add function to add a new object to the queue in case of creating a resource
func (op *Operator) Add(obj interface{}) {
	(*op.prometheusMiddleware.Gauge).Inc()
	op.OnResourceChange(obj, watch.Added)
}

// Delete function to add an object to the queue in case of deleting a resource
func (op *Operator) Delete(old interface{}) {
	(*op.prometheusMiddleware.Gauge).Dec()
	op.OnResourceChange(old, watch.Deleted)
}

func (op *Operator) OnResourceChange(obj interface{}, eventType watch.EventType) {
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

//Run function for controller which handles the queue
func (op *Operator) Run() {
	stopper := make(chan struct{})
	defer close(stopper)
	defer runtime.HandleCrash()

	go op.informer.Run(stopper)
	if !cache.WaitForCacheSync(stopper, op.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}
	<-stopper
}
