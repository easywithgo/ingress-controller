package ingress_controller

import (
	"context"
	"log"
	"sync"
	"time"

	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type RouterInfo struct {
	ingress *v1beta1.Ingress
	paths   []v1beta1.HTTPIngressPath
}

type IngressController struct {
	sync.Mutex
	ingress map[string]*v1beta1.Ingress
	router  map[string]RouterInfo
}

// InitIngressController initializes the Ingress controller.
// You need to provide the context which on cancel will stop the watcher for ingress updates
// and kubeClient which can be used to initialize the event listener
func InitIngressController(
	ctx context.Context,
	kubeClient *kubernetes.Clientset,
) *IngressController {
	ic := &IngressController{
		Mutex:   sync.Mutex{},
		ingress: make(map[string]*v1beta1.Ingress),
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go watchForUpdates(ctx, ic, kubeClient, wg)
	wg.Wait()
	return ic
}

// watchForUpdates continously looks for the ingress changes.
// It waits for maximum of 30 secs for the initial state to be set.
//
// 30 sec is chosen so to make the output of this function deterministic
func watchForUpdates(
	ctx context.Context,
	ic *IngressController,
	kubeClient *kubernetes.Clientset,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	listener := informers.
		NewSharedInformerFactory(kubeClient, time.Minute).
		Extensions().
		V1beta1().
		Ingresses().
		Informer()

	addToRouter := func(ingress *v1beta1.Ingress) {
		ic.Lock()
		defer ic.Unlock()
		paths := make([]v1beta1.HTTPIngressPath, 0)
		for _, rule := range ingress.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}
			paths = append(paths, rule.HTTP.Paths...)
			ic.router[rule.Host] = RouterInfo{
				ingress: ingress,
				paths:   paths,
			}
		}
	}

	deleteToRouter := func(ingress *v1beta1.Ingress) {
		for _, rule := range ingress.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}
			delete(ic.router, rule.Host)
		}
	}

	addToMap := func(key string, ingress *v1beta1.Ingress) {
		ic.Lock()
		defer ic.Unlock()
		ic.ingress[key] = ingress
	}

	removeFromMap := func(key string) {
		ic.Lock()
		defer ic.Unlock()
		delete(ic.ingress, key)
	}

	listener.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ingress, ok := obj.(*v1beta1.Ingress)
			if !ok {
				return
			}
			addToMap(ingress.Namespace+"/"+ingress.Name, ingress)
			addToRouter(ingress)
		},
		UpdateFunc: func(old, new interface{}) {
			o, ok := old.(*v1beta1.Ingress)
			if !ok {
				return
			}

			n, ok := new.(*v1beta1.Ingress)
			if !ok {
				return
			}
			removeFromMap(o.Namespace + "/" + o.Name)
			deleteToRouter(o)
			addToMap(n.Namespace+"/"+n.Name, n)
			addToRouter(n)
		},
		DeleteFunc: func(obj interface{}) {
			o, ok := obj.(*v1beta1.Ingress)
			if !ok {
				return
			}
			removeFromMap(o.Namespace + "/" + o.Name)
			deleteToRouter(o)
		},
	})

	go listener.Run(ctx.Done())
	iterator := 0
	for listener.HasSynced() || iterator >= 30 {
		time.Sleep(time.Second)
		iterator += 1
		log.Println("waiting for the data to be synced for first time")
	}
}
