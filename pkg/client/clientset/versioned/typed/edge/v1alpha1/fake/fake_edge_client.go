/*
Copyright The KCP Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"

	v1alpha1 "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/typed/edge/v1alpha1"
)

type FakeEdgeV1alpha1 struct {
	*testing.Fake
}

func (c *FakeEdgeV1alpha1) Customizers(namespace string) v1alpha1.CustomizerInterface {
	return &FakeCustomizers{c, namespace}
}

func (c *FakeEdgeV1alpha1) EdgePlacements() v1alpha1.EdgePlacementInterface {
	return &FakeEdgePlacements{c}
}

func (c *FakeEdgeV1alpha1) SinglePlacementSlices() v1alpha1.SinglePlacementSliceInterface {
	return &FakeSinglePlacementSlices{c}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeEdgeV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
