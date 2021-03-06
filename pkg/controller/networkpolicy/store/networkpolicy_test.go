// Copyright 2019 Antrea Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/vmware-tanzu/antrea/pkg/apis/networkpolicy"
	"github.com/vmware-tanzu/antrea/pkg/apiserver/storage"
	"github.com/vmware-tanzu/antrea/pkg/controller/types"
)

func TestWatchNetworkPolicyEvent(t *testing.T) {
	protocolTCP := networkpolicy.ProtocolTCP
	policyV1 := &types.NetworkPolicy{
		Namespace: "foo",
		Name:      "bar",
		SpanMeta:  types.SpanMeta{sets.NewString("node1", "node2")},
		Rules: []networkpolicy.NetworkPolicyRule{{
			Direction: networkpolicy.DirectionIn,
			From:      networkpolicy.NetworkPolicyPeer{AddressGroups: []string{"addressGroup1"}},
			To:        networkpolicy.NetworkPolicyPeer{},
			Services:  []networkpolicy.Service{{Protocol: &protocolTCP}},
		}},
		AppliedToGroups: []string{"appliedToGroup1"},
	}
	policyV2 := &types.NetworkPolicy{
		Namespace: "foo",
		Name:      "bar",
		SpanMeta:  types.SpanMeta{sets.NewString("node1", "node3")},
		Rules: []networkpolicy.NetworkPolicyRule{{
			Direction: networkpolicy.DirectionIn,
			From:      networkpolicy.NetworkPolicyPeer{AddressGroups: []string{"addressGroup1"}},
			To:        networkpolicy.NetworkPolicyPeer{},
			Services:  []networkpolicy.Service{{Protocol: &protocolTCP}},
		}},
		AppliedToGroups: []string{"appliedToGroup1"},
	}
	policyV3 := &types.NetworkPolicy{
		Namespace: "foo",
		Name:      "bar",
		SpanMeta:  types.SpanMeta{sets.NewString("node1", "node3")},
		Rules: []networkpolicy.NetworkPolicyRule{{
			Direction: networkpolicy.DirectionIn,
			From:      networkpolicy.NetworkPolicyPeer{AddressGroups: []string{"addressGroup2"}},
			To:        networkpolicy.NetworkPolicyPeer{},
			Services:  []networkpolicy.Service{{Protocol: &protocolTCP}},
		}},
		AppliedToGroups: []string{"appliedToGroup1"},
	}

	testCases := map[string]struct {
		fieldSelector fields.Selector
		// The operations that will be executed on the store.
		operations func(p storage.Interface)
		// The events expected to see.
		expected []watch.Event
	}{
		"non-node-scoped-watcher": {
			// All events should be watched.
			fieldSelector: fields.Everything(),
			operations: func(store storage.Interface) {
				store.Create(policyV1)
				store.Update(policyV2)
			},
			expected: []watch.Event{
				{watch.Added, &networkpolicy.NetworkPolicy{
					ObjectMeta:      metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
					Rules:           policyV1.Rules,
					AppliedToGroups: policyV1.AppliedToGroups,
				}},
				{watch.Modified, &networkpolicy.NetworkPolicy{
					ObjectMeta:      metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
					Rules:           policyV2.Rules,
					AppliedToGroups: policyV2.AppliedToGroups,
				}},
			},
		},
		"node-scoped-watcher": {
			// Only events that span node3 should be watched.
			fieldSelector: fields.SelectorFromSet(fields.Set{"nodeName": "node3"}),
			operations: func(store storage.Interface) {
				// This should not be seen as it doesn't span node3.
				store.Create(policyV1)
				// This should be seen as an added event as it makes the policy span node3 for the first time.
				store.Update(policyV2)
				// This should be seen as a modified event as it updates networkpolicies of node3.
				store.Update(policyV3)
				// This should be seen as a deleted event as it makes the policy not span node3 any more.
				store.Update(policyV1)
			},
			expected: []watch.Event{
				{watch.Added, &networkpolicy.NetworkPolicy{
					ObjectMeta:      metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
					Rules:           policyV2.Rules,
					AppliedToGroups: policyV2.AppliedToGroups,
				}},
				{watch.Modified, &networkpolicy.NetworkPolicy{
					ObjectMeta:      metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
					Rules:           policyV3.Rules,
					AppliedToGroups: policyV3.AppliedToGroups,
				}},
				{watch.Deleted, &networkpolicy.NetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{Namespace: "foo", Name: "bar"},
				}},
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			store := NewNetworkPolicyStore()
			w, err := store.Watch(context.Background(), "", labels.Everything(), testCase.fieldSelector)
			if err != nil {
				t.Fatalf("Failed to watch object: %v", err)
			}
			testCase.operations(store)
			ch := w.ResultChan()
			for _, expectedEvent := range testCase.expected {
				actualEvent := <-ch
				if !assert.Equal(t, expectedEvent, actualEvent) {
					t.Errorf("Expected event %v, got %v", expectedEvent, actualEvent)
				}
			}
			select {
			case obj, ok := <-ch:
				t.Errorf("Unexpected excess event: %v %t", obj, ok)
			default:
			}
		})
	}
}

func TestGetNetworkPolicyByIndex(t *testing.T) {
	policy1 := &types.NetworkPolicy{
		Namespace: "foo",
		Name:      "bar",
		Rules: []networkpolicy.NetworkPolicyRule{{
			Direction: networkpolicy.DirectionIn,
			From:      networkpolicy.NetworkPolicyPeer{AddressGroups: []string{"addressGroup1"}},
			To:        networkpolicy.NetworkPolicyPeer{},
		}},
		AppliedToGroups: []string{"appliedToGroup1"},
	}
	policy2 := &types.NetworkPolicy{
		Namespace: "foo2",
		Name:      "bar2",
		Rules: []networkpolicy.NetworkPolicyRule{{
			Direction: networkpolicy.DirectionIn,
			From:      networkpolicy.NetworkPolicyPeer{AddressGroups: []string{"addressGroup1", "addressGroup2"}},
			To:        networkpolicy.NetworkPolicyPeer{},
		}},
		AppliedToGroups: []string{"appliedToGroup1", "appliedToGroup2"},
	}

	testCases := map[string]struct {
		// The stored objects.
		networkPolicies []*types.NetworkPolicy
		// The index name used to get.
		indexName string
		// The index key used to get.
		indexKey string
		// The objects expected to be got by the indexName and indexKey.
		expectedNetworkPolicies []*types.NetworkPolicy
	}{
		"get-zero-by-addressgroup": {
			networkPolicies:         []*types.NetworkPolicy{policy1, policy2},
			indexName:               AddressGroupIndex,
			indexKey:                "non-existing-addressGroup",
			expectedNetworkPolicies: []*types.NetworkPolicy{},
		},
		"get-one-by-addressgroup": {
			networkPolicies:         []*types.NetworkPolicy{policy1, policy2},
			indexName:               AddressGroupIndex,
			indexKey:                "addressGroup2",
			expectedNetworkPolicies: []*types.NetworkPolicy{policy2},
		},
		"get-two-by-addressgroup": {
			networkPolicies:         []*types.NetworkPolicy{policy1, policy2},
			indexName:               AddressGroupIndex,
			indexKey:                "addressGroup1",
			expectedNetworkPolicies: []*types.NetworkPolicy{policy1, policy2},
		},
		"get-zero-by-appliedtogroup": {
			networkPolicies:         []*types.NetworkPolicy{policy1, policy2},
			indexName:               AppliedToGroupIndex,
			indexKey:                "non-existing-appliedToGroup",
			expectedNetworkPolicies: []*types.NetworkPolicy{},
		},
		"get-one-by-appliedtogroup": {
			networkPolicies:         []*types.NetworkPolicy{policy1, policy2},
			indexName:               AppliedToGroupIndex,
			indexKey:                "appliedToGroup2",
			expectedNetworkPolicies: []*types.NetworkPolicy{policy2},
		},
		"get-two-by-appliedtogroup": {
			networkPolicies:         []*types.NetworkPolicy{policy1, policy2},
			indexName:               AppliedToGroupIndex,
			indexKey:                "appliedToGroup1",
			expectedNetworkPolicies: []*types.NetworkPolicy{policy1, policy2},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			store := NewNetworkPolicyStore()
			for _, policy := range testCase.networkPolicies {
				if err := store.Create(policy); err != nil {
					t.Fatalf("Failed to store policy %v: %v", policy, err)
				}
			}

			actualNetworkPolicies, err := store.GetByIndex(testCase.indexName, testCase.indexKey)
			if err != nil {
				t.Fatalf("Failed to get policies by index %s/%s: %v", testCase.indexName, testCase.indexKey, err)
			}
			if !assert.ElementsMatch(t, testCase.expectedNetworkPolicies, actualNetworkPolicies) {
				t.Errorf("Expected policies %v, got %v", testCase.expectedNetworkPolicies, actualNetworkPolicies)
			}
		})
	}
}
