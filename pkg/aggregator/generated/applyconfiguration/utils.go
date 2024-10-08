// SPDX-License-Identifier: AGPL-3.0-only

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package applyconfiguration

import (
	v0alpha1 "github.com/grafana/grafana/pkg/aggregator/apis/aggregation/v0alpha1"
	aggregationv0alpha1 "github.com/grafana/grafana/pkg/aggregator/generated/applyconfiguration/aggregation/v0alpha1"
	internal "github.com/grafana/grafana/pkg/aggregator/generated/applyconfiguration/internal"
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	testing "k8s.io/client-go/testing"
)

// ForKind returns an apply configuration type for the given GroupVersionKind, or nil if no
// apply configuration type exists for the given GroupVersionKind.
func ForKind(kind schema.GroupVersionKind) interface{} {
	switch kind {
	// Group=aggregation.grafana.app, Version=v0alpha1
	case v0alpha1.SchemeGroupVersion.WithKind("DataPlaneService"):
		return &aggregationv0alpha1.DataPlaneServiceApplyConfiguration{}
	case v0alpha1.SchemeGroupVersion.WithKind("DataPlaneServiceCondition"):
		return &aggregationv0alpha1.DataPlaneServiceConditionApplyConfiguration{}
	case v0alpha1.SchemeGroupVersion.WithKind("DataPlaneServiceSpec"):
		return &aggregationv0alpha1.DataPlaneServiceSpecApplyConfiguration{}
	case v0alpha1.SchemeGroupVersion.WithKind("DataPlaneServiceStatus"):
		return &aggregationv0alpha1.DataPlaneServiceStatusApplyConfiguration{}
	case v0alpha1.SchemeGroupVersion.WithKind("Service"):
		return &aggregationv0alpha1.ServiceApplyConfiguration{}

	}
	return nil
}

func NewTypeConverter(scheme *runtime.Scheme) *testing.TypeConverter {
	return &testing.TypeConverter{Scheme: scheme, TypeResolver: internal.Parser()}
}
