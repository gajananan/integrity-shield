// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1alpha1

import (
	pkix "github.com/IBM/integrity-enforcer/develop/signservice/signservice-operator/pkg/pkix"
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SignService) DeepCopyInto(out *SignService) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SignService.
func (in *SignService) DeepCopy() *SignService {
	if in == nil {
		return nil
	}
	out := new(SignService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SignService) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SignServiceContainer) DeepCopyInto(out *SignServiceContainer) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SignServiceContainer.
func (in *SignServiceContainer) DeepCopy() *SignServiceContainer {
	if in == nil {
		return nil
	}
	out := new(SignServiceContainer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SignServiceList) DeepCopyInto(out *SignServiceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SignService, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SignServiceList.
func (in *SignServiceList) DeepCopy() *SignServiceList {
	if in == nil {
		return nil
	}
	out := new(SignServiceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SignServiceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SignServiceSpec) DeepCopyInto(out *SignServiceSpec) {
	*out = *in
	if in.ImagePullSecrets != nil {
		in, out := &in.ImagePullSecrets, &out.ImagePullSecrets
		*out = make([]v1.LocalObjectReference, len(*in))
		copy(*out, *in)
	}
	in.SignService.DeepCopyInto(&out.SignService)
	if in.Signers != nil {
		in, out := &in.Signers, &out.Signers
		*out = make([]pkix.SignerCertName, len(*in))
		copy(*out, *in)
	}
	if in.InvalidSigners != nil {
		in, out := &in.InvalidSigners, &out.InvalidSigners
		*out = make([]pkix.SignerCertName, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SignServiceSpec.
func (in *SignServiceSpec) DeepCopy() *SignServiceSpec {
	if in == nil {
		return nil
	}
	out := new(SignServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SignServiceStatus) DeepCopyInto(out *SignServiceStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SignServiceStatus.
func (in *SignServiceStatus) DeepCopy() *SignServiceStatus {
	if in == nil {
		return nil
	}
	out := new(SignServiceStatus)
	in.DeepCopyInto(out)
	return out
}
