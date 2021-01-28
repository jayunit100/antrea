// Copyright 2020 Antrea Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by client-gen. DO NOT EDIT.

package v1beta2

import (
	"context"
	"time"

	v1beta2 "github.com/vmware-tanzu/antrea/pkg/apis/controlplane/v1beta2"
	scheme "github.com/vmware-tanzu/antrea/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// GroupsGetter has a method to return a GroupInterface.
// A group's client should implement this interface.
type GroupsGetter interface {
	Groups() GroupInterface
}

// GroupInterface has methods to work with Group resources.
type GroupInterface interface {
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1beta2.Group, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1beta2.GroupList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	GroupExpansion
}

// groups implements GroupInterface
type groups struct {
	client rest.Interface
}

// newGroups returns a Groups
func newGroups(c *ControlplaneV1beta2Client) *groups {
	return &groups{
		client: c.RESTClient(),
	}
}

// Get takes name of the group, and returns the corresponding group object, and an error if there is any.
func (c *groups) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta2.Group, err error) {
	result = &v1beta2.Group{}
	err = c.client.Get().
		Resource("groups").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Groups that match those selectors.
func (c *groups) List(ctx context.Context, opts v1.ListOptions) (result *v1beta2.GroupList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1beta2.GroupList{}
	err = c.client.Get().
		Resource("groups").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested groups.
func (c *groups) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("groups").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}