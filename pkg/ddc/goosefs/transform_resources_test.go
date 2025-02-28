/*
Copyright 2022 The Fluid Authors.

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

package goosefs

import (
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestTransformResourcesForMaster(t *testing.T) {
	testCases := map[string]struct {
		runtime *datav1alpha1.GooseFSRuntime
		got     *GooseFS
		want    *GooseFS
	}{
		"test goosefs master pass through resources with limits and request case 1": {
			runtime: mockGooseFSRuntimeForMaster(
				corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("100Mi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("400m"),
						corev1.ResourceMemory: resource.MustParse("400Mi"),
					},
				},
			),
			got: &GooseFS{},
			want: &GooseFS{
				Master: Master{
					Resources: common.Resources{
						Requests: common.ResourceList{
							corev1.ResourceCPU:    "100m",
							corev1.ResourceMemory: "100Mi",
						},
						Limits: common.ResourceList{
							corev1.ResourceCPU:    "400m",
							corev1.ResourceMemory: "400Mi",
						},
					},
				},
				JobMaster: JobMaster{
					Resources: common.Resources{
						Requests: common.ResourceList{
							corev1.ResourceCPU:    "100m",
							corev1.ResourceMemory: "100Mi",
						},
						Limits: common.ResourceList{
							corev1.ResourceCPU:    "400m",
							corev1.ResourceMemory: "400Mi",
						},
					},
				},
			},
		},
		"test GooseFS master pass through resources with request case 1": {
			runtime: mockGooseFSRuntimeForMaster(
				corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("100Mi"),
					},
				},
			),
			got: &GooseFS{},
			want: &GooseFS{
				Master: Master{
					Resources: common.Resources{
						Requests: common.ResourceList{
							corev1.ResourceCPU:    "100m",
							corev1.ResourceMemory: "100Mi",
						},
					},
				},
				JobMaster: JobMaster{
					Resources: common.Resources{
						Requests: common.ResourceList{
							corev1.ResourceCPU:    "100m",
							corev1.ResourceMemory: "100Mi",
						},
					},
				},
			},
		},
		"test goosefs master pass through resources without request and limit case 1": {
			runtime: mockGooseFSRuntimeForMaster(
				corev1.ResourceRequirements{
					Requests: corev1.ResourceList{},
				},
			),
			got:  &GooseFS{},
			want: &GooseFS{},
		},
		"test goosefs master pass through resources without request and limit case 2": {
			runtime: mockGooseFSRuntimeForMaster(corev1.ResourceRequirements{}),
			got:     &GooseFS{},
			want:    &GooseFS{},
		},
		"test goosefs master pass through resources without request and limit case 3": {
			runtime: mockGooseFSRuntimeForMaster(
				corev1.ResourceRequirements{
					Limits: corev1.ResourceList{},
				},
			),
			got:  &GooseFS{},
			want: &GooseFS{},
		},
	}

	engine := &GooseFSEngine{}
	for k, item := range testCases {
		engine.transformResourcesForMaster(item.runtime, item.got)
		if !reflect.DeepEqual(item.want.Master.Resources, item.got.Master.Resources) {
			t.Errorf("%s failure, want resource: %+v,got resource: %+v",
				k,
				item.want.Master.Resources,
				item.got.Master.Resources,
			)
		}
	}
}

func mockGooseFSRuntimeForMaster(res corev1.ResourceRequirements) *datav1alpha1.GooseFSRuntime {
	runtime := &datav1alpha1.GooseFSRuntime{
		Spec: datav1alpha1.GooseFSRuntimeSpec{
			Master: datav1alpha1.GooseFSCompTemplateSpec{
				Resources: res,
			},
			JobMaster: datav1alpha1.GooseFSCompTemplateSpec{
				Resources: res,
			},
		},
	}
	return runtime

}

func TestTransformResourcesForWorkerNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{},
		}, &GooseFS{
			Properties: map[string]string{},
		}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{Log: log.NullLogger{}}
		engine.transformResourcesForWorker(test.runtime, test.goosefsValue)
		if result, found := test.goosefsValue.Worker.Resources.Limits[corev1.ResourceMemory]; found {
			t.Errorf("expected nil, got %v", result)
		}
	}
}

func TestTransformResourcesForWorkerWithValue(t *testing.T) {

	resources := corev1.ResourceRequirements{}
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("2Gi")
	resources.Limits[corev1.ResourceCPU] = resource.MustParse("500m")
	resources.Requests = make(corev1.ResourceList)
	resources.Requests[corev1.ResourceMemory] = resource.MustParse("1Gi")
	resources.Requests[corev1.ResourceCPU] = resource.MustParse("500m")

	result := resource.MustParse("20Gi")

	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Worker: datav1alpha1.GooseFSCompTemplateSpec{
					Resources: resources,
				},
				JobWorker: datav1alpha1.GooseFSCompTemplateSpec{
					Resources: resources,
				},
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
					}},
				},
			},
		}, &GooseFS{
			Properties: map[string]string{},
			Master:     Master{},
		}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{Log: log.NullLogger{}}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "goosefs", test.runtime.Spec.TieredStore)
		engine.UnitTest = true
		engine.transformResourcesForWorker(test.runtime, test.goosefsValue)
		if test.goosefsValue.Worker.Resources.Limits[corev1.ResourceMemory] != "22Gi" {
			t.Errorf("expected 22Gi, got %v", test.goosefsValue.Worker.Resources.Limits[corev1.ResourceMemory])
		}
	}
}

func TestTransformResourcesForFuseNoValue(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{},
		}, &GooseFS{
			Properties: map[string]string{},
		}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{Log: log.NullLogger{}}
		engine.transformResourcesForFuse(test.runtime, test.goosefsValue)
		if result, found := test.goosefsValue.Fuse.Resources.Limits[corev1.ResourceMemory]; found {
			t.Errorf("expected nil, got %v", result)
		}
	}
}

func TestTransformResourcesForFuseWithValue(t *testing.T) {

	resources := corev1.ResourceRequirements{}
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("2Gi")

	result := resource.MustParse("20Gi")

	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Fuse: datav1alpha1.GooseFSFuseSpec{
					Resources: resources,
				},
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
					}},
				},
			},
		}, &GooseFS{
			Properties: map[string]string{},
			Master:     Master{},
			JobMaster:  JobMaster{},
		}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{Log: log.NullLogger{}}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo("test", "test", "goosefs", test.runtime.Spec.TieredStore)
		engine.UnitTest = true
		engine.transformResourcesForFuse(test.runtime, test.goosefsValue)
		if test.goosefsValue.Fuse.Resources.Limits[corev1.ResourceMemory] != "22Gi" {
			t.Errorf("expected 22Gi, got %v", test.goosefsValue.Fuse.Resources.Limits[corev1.ResourceMemory])
		}
	}
}
