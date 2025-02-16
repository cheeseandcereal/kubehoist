//go:build !ignore_autogenerated

/*
Copyright 2025.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControllerWatch) DeepCopyInto(out *ControllerWatch) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControllerWatch.
func (in *ControllerWatch) DeepCopy() *ControllerWatch {
	if in == nil {
		return nil
	}
	out := new(ControllerWatch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ControllerWatch) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControllerWatchList) DeepCopyInto(out *ControllerWatchList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ControllerWatch, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControllerWatchList.
func (in *ControllerWatchList) DeepCopy() *ControllerWatchList {
	if in == nil {
		return nil
	}
	out := new(ControllerWatchList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ControllerWatchList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControllerWatchSpec) DeepCopyInto(out *ControllerWatchSpec) {
	*out = *in
	in.HelmControllerSpec.DeepCopyInto(&out.HelmControllerSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControllerWatchSpec.
func (in *ControllerWatchSpec) DeepCopy() *ControllerWatchSpec {
	if in == nil {
		return nil
	}
	out := new(ControllerWatchSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControllerWatchStatus) DeepCopyInto(out *ControllerWatchStatus) {
	*out = *in
	if in.InstalledCRDs != nil {
		in, out := &in.InstalledCRDs, &out.InstalledCRDs
		*out = make([]GroupVersionKind, len(*in))
		copy(*out, *in)
	}
	if in.LastUpdated != nil {
		in, out := &in.LastUpdated, &out.LastUpdated
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControllerWatchStatus.
func (in *ControllerWatchStatus) DeepCopy() *ControllerWatchStatus {
	if in == nil {
		return nil
	}
	out := new(ControllerWatchStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GroupVersionKind) DeepCopyInto(out *GroupVersionKind) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GroupVersionKind.
func (in *GroupVersionKind) DeepCopy() *GroupVersionKind {
	if in == nil {
		return nil
	}
	out := new(GroupVersionKind)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HelmInstallSpec) DeepCopyInto(out *HelmInstallSpec) {
	*out = *in
	if in.CreateNamespace != nil {
		in, out := &in.CreateNamespace, &out.CreateNamespace
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HelmInstallSpec.
func (in *HelmInstallSpec) DeepCopy() *HelmInstallSpec {
	if in == nil {
		return nil
	}
	out := new(HelmInstallSpec)
	in.DeepCopyInto(out)
	return out
}
