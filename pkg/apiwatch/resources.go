/*
Copyright 2023 The KubeStellar Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package apiwatch

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	upstreamdiscovery "k8s.io/client-go/discovery"
	cachediscovery "k8s.io/client-go/discovery/cached/memory"
	upstreamcache "k8s.io/client-go/tools/cache"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	"k8s.io/klog/v2"

	ksmetav1a1 "github.com/kubestellar/kubestellar/pkg/apis/meta/v1alpha1"
)

// Invalidatable is a cache that has to be explicitly invalidated
type Invalidatable interface {
	// Invalidate the cache
	Invalidate()
}

// ObjectNotifier is something that notifies the client like an informer does
type ObjectNotifier interface {
	AddEventHandler(handler upstreamcache.ResourceEventHandler)
}

type ResourceDefinitionSupplier interface {
	ObjectNotifier
	GetGVK(obj any) schema.GroupVersionKind
	EnumerateDefinedResources(definer any) ResourceDefinitionEnumerator
}

type ResourceDefinitionEnumerator func(func(metav1.GroupVersionResource))

// APIResourceLister helps list APIResources.
// All objects returned here must be treated as read-only.
type APIResourceLister interface {
	// List lists all APIResources in the informer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*ksmetav1a1.APIResource, err error)
	// Get retrieves the APIResource having the given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*ksmetav1a1.APIResource, error)
}

// NewAPIResourceInformer creates an informer on the API resources
// revealed by the given client.  The objects delivered by the
// informer are of type `*ksmetav1a1.APIResource`.
//
// The results from the given client are cached in memory and that
// cache has to be explicitly invalidated.  Invalidation can be done
// by calling the returned Invalidator.  Additionally, invalidation
// happens whenever any of the supplied invalidationNotifiers delivers
// a notification of an object addition.  Re-querying the given client
// is delayed by a few decaseconds (with Nagling) to support
// invalidations based on events that merely trigger some process of
// changing the set of API resources.
func NewAPIResourceInformer(ctx context.Context, clusterName string, client upstreamdiscovery.DiscoveryInterface, includeSubresources bool, invalidationNotifiers ...ObjectNotifier) (upstreamcache.SharedInformer, APIResourceLister, Invalidatable) {
	logger := klog.FromContext(ctx).WithValues("cluster", clusterName)
	ctx = klog.NewContext(ctx, logger)
	rlw := &resourcesListWatcher{
		ctx:                 ctx,
		logger:              logger,
		includeSubresources: includeSubresources,
		clusterName:         clusterName,
		cache:               cachediscovery.NewMemCacheClient(client),
		resourceVersionI:    1,
		rscToDefiners:       GoMap[metav1.GroupVersionResource, GoSet[objectID]]{},
		definerToRscs:       GoMap[objectID, GoSet[metav1.GroupVersionResource]]{},
	}
	rlw.cond = sync.NewCond(&rlw.mutex)
	go func() {
		doneCh := ctx.Done()
		for {
			select {
			case <-doneCh:
				return
			default:
			}
			var wait time.Duration
			func() {
				rlw.mutex.Lock()
				defer rlw.mutex.Unlock()
				if rlw.needRelist {
					now := time.Now()
					if now.Before(rlw.relistAfter) {
						wait = rlw.relistAfter.Sub(now)
					} else {
						logger.V(3).Info("Cycled APIResourceInformer")
						for _, cancel := range rlw.cancels {
							cancel()
						}
						rlw.needRelist = false
					}
					return
				}
				rlw.cond.Wait()
			}()
			if wait > 0 {
				time.Sleep(wait)
			}
		}
	}()
	for _, invalidator := range invalidationNotifiers {
		supplier, isSupplier := invalidator.(ResourceDefinitionSupplier)
		invalidator.AddEventHandler(upstreamcache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				logger.V(3).Info("Notified of invalidator", "obj", obj, "isSupplier", isSupplier)
				rlw.InvalidateWithDefiner(obj, supplier, true)
			},
			UpdateFunc: func(oldObj, newObj any) {
				logger.V(3).Info("Notified of invalidator change", "newObj", newObj, "isSupplier", isSupplier)
				rlw.InvalidateWithDefiner(newObj, supplier, true)
			},
			DeleteFunc: func(obj any) {
				if del, ok := obj.(upstreamcache.DeletedFinalStateUnknown); ok {
					obj = del.Obj
				}
				logger.V(3).Info("Notified of invalidator deletion", "obj", obj, "isSupplier", isSupplier)
				rlw.InvalidateWithDefiner(obj, supplier, false)
			},
		})
	}
	inf := upstreamcache.NewSharedInformer(rlw, &ksmetav1a1.APIResource{}, 0)
	return inf, resourceLister{inf.GetStore()}, rlw
}

type resourcesListWatcher struct {
	ctx                 context.Context
	logger              klog.Logger
	includeSubresources bool
	clusterName         string
	cache               upstreamdiscovery.CachedDiscoveryInterface

	mutex            sync.Mutex
	cond             *sync.Cond
	resourceVersionI int64
	needRelist       bool
	relistAfter      time.Time
	cancels          []context.CancelFunc
	rscToDefiners    GoMap[metav1.GroupVersionResource, GoSet[objectID]]
	definerToRscs    GoMap[objectID, GoSet[metav1.GroupVersionResource]]
}

// objectID identifies an object that defines resources
type objectID struct {
	APIVersion string // including group
	Kind       string
	Name       string
}

type Empty struct{}

// GoMap is the usual with a MarshalJSON method
type GoMap[Key comparable, Val any] map[Key]Val

// GoSet is the usual with a MarshalJSON method
type GoSet[Key comparable] GoMap[Key, Empty]

var _ json.Marshaler = GoMap[int, func()]{}

func (this GoMap[Key, Val]) MarshalJSON() ([]byte, error) {
	return MarshalMap(this)
}

var _ json.Marshaler = GoSet[int]{}

func (this GoSet[Key]) MarshalJSON() ([]byte, error) {
	return MarshalSet(this)
}

func (rlw *resourcesListWatcher) InvalidateWithDefiner(obj any, supplier ResourceDefinitionSupplier, set bool) {
	rlw.mutex.Lock()
	defer rlw.mutex.Unlock()
	rlw.invalidateWithDefinerLocked(obj, supplier, set)
}

func (rlw *resourcesListWatcher) Invalidate() {
	rlw.mutex.Lock()
	defer rlw.mutex.Unlock()
	rlw.invalidateWithDefinerLocked(nil, nil, false)
}

func (rlw *resourcesListWatcher) invalidateWithDefinerLocked(obj any, supplier ResourceDefinitionSupplier, set bool) {
	rlw.resourceVersionI += 1
	rlw.relistAfter = time.Now().Add(time.Second * 20)
	rlw.needRelist = true
	rlw.cache.Invalidate()
	rlw.cond.Broadcast()
	if obj == nil || supplier == nil {
		return
	}
	objM := obj.(metav1.Object)
	gvk := supplier.GetGVK(obj)
	rlw.logger.V(4).Info("Examining resource definer", "obj", obj, "supplierType", fmt.Sprintf("%T", supplier), "gvk", gvk)
	apiVersion, kind := gvk.ToAPIVersionAndKind()
	oid := objectID{apiVersion, kind, objM.GetName()}
	if oid.APIVersion == "" {
		panic(obj)
	}
	if oid.Kind == "" {
		panic(obj)
	}
	var enumr ResourceDefinitionEnumerator = enumerateNothing
	if set {
		enumr = supplier.EnumerateDefinedResources(obj)
	}
	rlw.setDefinerLocked(oid, enumr)
}

func enumerateNothing(func(metav1.GroupVersionResource)) {}

type resourceWatch struct {
	*resourcesListWatcher
	cancel  context.CancelFunc
	results chan watch.Event
}

func (rw *resourceWatch) ResultChan() <-chan watch.Event {
	return rw.results
}

func (rw *resourceWatch) Stop() {
	rw.cancel()
}

func (rlw *resourcesListWatcher) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	rlw.mutex.Lock()
	defer rlw.mutex.Unlock()
	resourceVersionS := strconv.FormatInt(rlw.resourceVersionI, 10)
	if resourceVersionS != opts.ResourceVersion {
		return nil, apierrors.NewResourceExpired(fmt.Sprintf("Requested version %s, have version %s in cluster %s", opts.ResourceVersion, resourceVersionS, rlw.clusterName))
	}
	timeout := time.Duration(*opts.TimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(rlw.ctx, timeout)
	rw := &resourceWatch{
		resourcesListWatcher: rlw,
		cancel:               cancel,
		results:              make(chan watch.Event),
	}
	rlw.cancels = append(rlw.cancels, cancel)
	go func() {
		<-ctx.Done()
		rlw.logger.V(3).Info("Ending an APIResource Watch")
		close(rw.results)
	}()
	return rw, nil
}

func (rlw *resourcesListWatcher) List(opts metav1.ListOptions) (k8sruntime.Object, error) {
	resourceVersionI := func() int64 {
		rlw.mutex.Lock()
		defer rlw.mutex.Unlock()
		rlw.resourceVersionI = rlw.resourceVersionI + 1
		for _, cancel := range rlw.cancels {
			cancel()
		}
		return rlw.resourceVersionI
	}()
	resourceVersionS := strconv.FormatInt(resourceVersionI, 10)
	ans := ksmetav1a1.APIResourceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIResourceList",
			APIVersion: ksmetav1a1.SchemeGroupVersion.String(),
		},
		ListMeta: metav1.ListMeta{ResourceVersion: resourceVersionS},
	}
	var err error
	if rlw.includeSubresources {
		ans.Items, err = rlw.listWithSubresources(rlw.logger, resourceVersionS)
	} else {
		ans.Items, err = rlw.listSansSubresources(resourceVersionS)
	}
	return &ans, err
}

// arMap maps from resource or subresource name (single step in pathname) to data for that name
type arMap map[string]*arTuple

// arTuple holds the data for an APIResource
type arTuple struct {
	spec         *ksmetav1a1.APIResourceSpec
	subresources arMap
}

func (am arMap) insert(name []string, spec *ksmetav1a1.APIResourceSpec) {
	art := am[name[0]]
	if art == nil {
		art = &arTuple{subresources: arMap{}}
		am[name[0]] = art
	}
	if len(name) < 2 {
		art.spec = spec
	} else {
		art.subresources.insert(name[1:], spec)
	}
}

func (am arMap) toList(logger klog.Logger, prefix []string, consume func(ksmetav1a1.APIResourceSpec)) {
	for name, art := range am {
		if art.spec == nil {
			logger.Error(nil, "Gap in subresource structure", "prefix", prefix, "name", name, "subresources", art.subresources)
			continue
		}
		spec := *art.spec
		spec.Name = name
		art.subresources.toList(logger, append(prefix, name), func(subSpec ksmetav1a1.APIResourceSpec) {
			spec.SubResources = append(spec.SubResources, &subSpec)
		})
		consume(spec)
	}
}

func (rlw *resourcesListWatcher) listWithSubresources(logger klog.Logger, resourceVersionS string) ([]ksmetav1a1.APIResource, error) {
	groupList, resourceList, err := rlw.cache.ServerGroupsAndResources()
	if err != nil {
		rlw.logger.V(3).Info("Did not get all api groups and resources", "err", err.Error())
	}
	groupToVersion := map[string]string{}
	for _, ag := range groupList {
		groupToVersion[ag.Name] = ag.PreferredVersion.Version
	}
	ans := []ksmetav1a1.APIResource{}
	rlw.mutex.Lock()
	defer rlw.mutex.Unlock()
	for _, group := range resourceList {
		gv, err := schema.ParseGroupVersion(group.GroupVersion)
		if err != nil {
			rlw.logger.Error(err, "Failed to parse a GroupVersion", "groupVersion", group.GroupVersion)
			continue
		}
		if groupToVersion[gv.Group] != gv.Version {
			rlw.logger.V(4).Info("Ignoring wrong version", "gv", gv, "rightVersion", groupToVersion[gv.Group])
			continue
		}
		am := arMap{}
		rlw.enumAPIResourcesLocked(resourceVersionS, gv, group.APIResources, func(ar ksmetav1a1.APIResourceSpec) {
			rscName := ar.Name
			nameParts := strings.Split(rscName, "/")
			am.insert(nameParts, &ar)
		})
		am.toList(logger, []string{}, func(spec ksmetav1a1.APIResourceSpec) {
			complete := specComplete(spec, resourceVersionS, gv)
			ans = append(ans, complete)
		})
	}
	return ans, nil
}

func specComplete(spec ksmetav1a1.APIResourceSpec, resourceVersionS string, gv schema.GroupVersion) ksmetav1a1.APIResource {
	return ksmetav1a1.APIResource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIResource",
			APIVersion: ksmetav1a1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			// The normal syntax has a slash, which confuses the usual Store
			Name:            gv.Group + ":" + gv.Version + ":" + spec.Name,
			ResourceVersion: resourceVersionS,
		},
		Spec: spec}
}

func (rlw *resourcesListWatcher) enumAPIResourcesLocked(resourceVersionS string, gv schema.GroupVersion, mrs []metav1.APIResource, consumer func(ksmetav1a1.APIResourceSpec)) {
	for _, rsc := range mrs {
		rscVersion := rsc.Version
		if rscVersion == "" {
			rscVersion = gv.Version
		}
		gvr := metav1.GroupVersionResource{Group: gv.Group, Version: rscVersion, Resource: rsc.Name}
		definers := definersToSlice(rlw.rscToDefiners[gvr])
		rlw.logger.V(4).Info("Enumerating", "gvr", gvr, "definers", definers)
		arSpec := ksmetav1a1.APIResourceSpec{
			Name:         rsc.Name,
			SingularName: rsc.SingularName,
			Namespaced:   rsc.Namespaced,
			Group:        gv.Group,
			Version:      rscVersion,
			Kind:         rsc.Kind,
			Verbs:        rsc.Verbs,
			Definers:     definers,
		}
		// rlw.logger.V(4).Info("Producing an APIResource", "ar", ar)
		consumer(arSpec)
	}
}

func definersToSlice(asSet map[objectID]Empty) []ksmetav1a1.Definer {
	ans := make([]ksmetav1a1.Definer, 0, len(asSet))
	for definer := range asSet {
		ans = append(ans, ksmetav1a1.Definer{Kind: definer.Kind, Name: definer.Name})
	}
	return ans
}

func (rlw *resourcesListWatcher) listSansSubresources(resourceVersionS string) ([]ksmetav1a1.APIResource, error) {
	groupList, err := rlw.cache.ServerPreferredResources()
	if err != nil {
		rlw.logger.V(3).Info("Did not get all preferred resources", "err", err.Error())
	}
	ans := []ksmetav1a1.APIResource{}
	rlw.mutex.Lock()
	defer rlw.mutex.Unlock()
	for _, group := range groupList {
		gv, err := schema.ParseGroupVersion(group.GroupVersion)
		if err != nil {
			rlw.logger.Error(err, "Failed to parse a GroupVersion", "groupVersion", group.GroupVersion)
			continue
		}
		rlw.enumAPIResourcesLocked(resourceVersionS, gv, group.APIResources, func(arSpec ksmetav1a1.APIResourceSpec) {
			ar := specComplete(arSpec, resourceVersionS, gv)
			ans = append(ans, ar)
		})
	}
	return ans, nil
}

type resourceLister struct {
	store upstreamcache.Store
}

func (rl resourceLister) List(selector labels.Selector) (ret []*ksmetav1a1.APIResource, err error) {
	allObjs := rl.store.List()
	for _, obj := range allObjs {
		ar := obj.(*ksmetav1a1.APIResource)
		if selector.Matches(labels.Set(ar.Labels)) {
			ret = append(ret, ar)
		}
	}
	return
}

func (rl resourceLister) Get(name string) (*ksmetav1a1.APIResource, error) {
	obj, exists, err := rl.store.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		gr := schema.GroupResource{Group: ksmetav1a1.SchemeGroupVersion.Group, Resource: "apiresources"}
		return nil, apierrors.NewNotFound(gr, name)
	}
	return obj.(*ksmetav1a1.APIResource), nil
}
