//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright The KubeStellar Authors.

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

// Code generated by kcp code-generator. DO NOT EDIT.

package v2alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/testing"

	kcptesting "github.com/kcp-dev/client-go/third_party/k8s.io/client-go/testing"
	"github.com/kcp-dev/logicalcluster/v3"

	edgev2alpha1 "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	edgev2alpha1client "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned/typed/edge/v2alpha1"
)

var edgePlacementDecisionsResource = schema.GroupVersionResource{Group: "edge.kubestellar.io", Version: "v2alpha1", Resource: "edgeplacementdecisions"}
var edgePlacementDecisionsKind = schema.GroupVersionKind{Group: "edge.kubestellar.io", Version: "v2alpha1", Kind: "EdgePlacementDecision"}

type edgePlacementDecisionsClusterClient struct {
	*kcptesting.Fake
}

// Cluster scopes the client down to a particular cluster.
func (c *edgePlacementDecisionsClusterClient) Cluster(clusterPath logicalcluster.Path) edgev2alpha1client.EdgePlacementDecisionInterface {
	if clusterPath == logicalcluster.Wildcard {
		panic("A specific cluster must be provided when scoping, not the wildcard.")
	}

	return &edgePlacementDecisionsClient{Fake: c.Fake, ClusterPath: clusterPath}
}

// List takes label and field selectors, and returns the list of EdgePlacementDecisions that match those selectors across all clusters.
func (c *edgePlacementDecisionsClusterClient) List(ctx context.Context, opts metav1.ListOptions) (*edgev2alpha1.EdgePlacementDecisionList, error) {
	obj, err := c.Fake.Invokes(kcptesting.NewRootListAction(edgePlacementDecisionsResource, edgePlacementDecisionsKind, logicalcluster.Wildcard, opts), &edgev2alpha1.EdgePlacementDecisionList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &edgev2alpha1.EdgePlacementDecisionList{ListMeta: obj.(*edgev2alpha1.EdgePlacementDecisionList).ListMeta}
	for _, item := range obj.(*edgev2alpha1.EdgePlacementDecisionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested EdgePlacementDecisions across all clusters.
func (c *edgePlacementDecisionsClusterClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.InvokesWatch(kcptesting.NewRootWatchAction(edgePlacementDecisionsResource, logicalcluster.Wildcard, opts))
}

type edgePlacementDecisionsClient struct {
	*kcptesting.Fake
	ClusterPath logicalcluster.Path
}

func (c *edgePlacementDecisionsClient) Create(ctx context.Context, edgePlacementDecision *edgev2alpha1.EdgePlacementDecision, opts metav1.CreateOptions) (*edgev2alpha1.EdgePlacementDecision, error) {
	obj, err := c.Fake.Invokes(kcptesting.NewRootCreateAction(edgePlacementDecisionsResource, c.ClusterPath, edgePlacementDecision), &edgev2alpha1.EdgePlacementDecision{})
	if obj == nil {
		return nil, err
	}
	return obj.(*edgev2alpha1.EdgePlacementDecision), err
}

func (c *edgePlacementDecisionsClient) Update(ctx context.Context, edgePlacementDecision *edgev2alpha1.EdgePlacementDecision, opts metav1.UpdateOptions) (*edgev2alpha1.EdgePlacementDecision, error) {
	obj, err := c.Fake.Invokes(kcptesting.NewRootUpdateAction(edgePlacementDecisionsResource, c.ClusterPath, edgePlacementDecision), &edgev2alpha1.EdgePlacementDecision{})
	if obj == nil {
		return nil, err
	}
	return obj.(*edgev2alpha1.EdgePlacementDecision), err
}

func (c *edgePlacementDecisionsClient) UpdateStatus(ctx context.Context, edgePlacementDecision *edgev2alpha1.EdgePlacementDecision, opts metav1.UpdateOptions) (*edgev2alpha1.EdgePlacementDecision, error) {
	obj, err := c.Fake.Invokes(kcptesting.NewRootUpdateSubresourceAction(edgePlacementDecisionsResource, c.ClusterPath, "status", edgePlacementDecision), &edgev2alpha1.EdgePlacementDecision{})
	if obj == nil {
		return nil, err
	}
	return obj.(*edgev2alpha1.EdgePlacementDecision), err
}

func (c *edgePlacementDecisionsClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.Invokes(kcptesting.NewRootDeleteActionWithOptions(edgePlacementDecisionsResource, c.ClusterPath, name, opts), &edgev2alpha1.EdgePlacementDecision{})
	return err
}

func (c *edgePlacementDecisionsClient) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := kcptesting.NewRootDeleteCollectionAction(edgePlacementDecisionsResource, c.ClusterPath, listOpts)

	_, err := c.Fake.Invokes(action, &edgev2alpha1.EdgePlacementDecisionList{})
	return err
}

func (c *edgePlacementDecisionsClient) Get(ctx context.Context, name string, options metav1.GetOptions) (*edgev2alpha1.EdgePlacementDecision, error) {
	obj, err := c.Fake.Invokes(kcptesting.NewRootGetAction(edgePlacementDecisionsResource, c.ClusterPath, name), &edgev2alpha1.EdgePlacementDecision{})
	if obj == nil {
		return nil, err
	}
	return obj.(*edgev2alpha1.EdgePlacementDecision), err
}

// List takes label and field selectors, and returns the list of EdgePlacementDecisions that match those selectors.
func (c *edgePlacementDecisionsClient) List(ctx context.Context, opts metav1.ListOptions) (*edgev2alpha1.EdgePlacementDecisionList, error) {
	obj, err := c.Fake.Invokes(kcptesting.NewRootListAction(edgePlacementDecisionsResource, edgePlacementDecisionsKind, c.ClusterPath, opts), &edgev2alpha1.EdgePlacementDecisionList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &edgev2alpha1.EdgePlacementDecisionList{ListMeta: obj.(*edgev2alpha1.EdgePlacementDecisionList).ListMeta}
	for _, item := range obj.(*edgev2alpha1.EdgePlacementDecisionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

func (c *edgePlacementDecisionsClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.InvokesWatch(kcptesting.NewRootWatchAction(edgePlacementDecisionsResource, c.ClusterPath, opts))
}

func (c *edgePlacementDecisionsClient) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*edgev2alpha1.EdgePlacementDecision, error) {
	obj, err := c.Fake.Invokes(kcptesting.NewRootPatchSubresourceAction(edgePlacementDecisionsResource, c.ClusterPath, name, pt, data, subresources...), &edgev2alpha1.EdgePlacementDecision{})
	if obj == nil {
		return nil, err
	}
	return obj.(*edgev2alpha1.EdgePlacementDecision), err
}
