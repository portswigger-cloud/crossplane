/*
Copyright 2022 The Crossplane Authors.

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

package v1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/crossplane/crossplane/apis/apiextensions/v1beta1"
)

func TestLatestRevision(t *testing.T) {
	ctrl := true

	comp := &Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cool-composition",
			UID:  types.UID("no-you-uid"),
			Labels: map[string]string{
				"channel": "dev",
			},
		},
	}

	// Owned by the above composition, with an old hash.
	rev1 := &v1beta1.CompositionRevision{
		ObjectMeta: metav1.ObjectMeta{
			Name: comp.GetName() + "-1",
			OwnerReferences: []metav1.OwnerReference{{
				UID:                comp.GetUID(),
				Controller:         &ctrl,
				BlockOwnerDeletion: &ctrl,
			}},
			Labels: map[string]string{
				v1beta1.LabelCompositionHash: "some-older-hash",
				v1beta1.LabelCompositionName: comp.Name,
			},
		},
		Spec: v1beta1.CompositionRevisionSpec{Revision: 1},
	}

	// Owned by the above composition, with the current hash.
	rev2 := &v1beta1.CompositionRevision{
		ObjectMeta: metav1.ObjectMeta{
			Name: comp.GetName() + "-2",
			OwnerReferences: []metav1.OwnerReference{{
				UID:                comp.GetUID(),
				Controller:         &ctrl,
				BlockOwnerDeletion: &ctrl,
			}},
			Labels: map[string]string{
				v1beta1.LabelCompositionHash: comp.Hash()[:63],
				v1beta1.LabelCompositionName: comp.Name,
			},
		},
		Spec: v1beta1.CompositionRevisionSpec{Revision: 2},
	}

	// Not owned by the above composition. Has the largest revision number.
	rev3 := &v1beta1.CompositionRevision{
		ObjectMeta: metav1.ObjectMeta{
			Name: comp.GetName() + "-3",
		},
		Spec: v1beta1.CompositionRevisionSpec{Revision: 3},
	}

	cases := map[string]struct {
		reason string
		args   []v1beta1.CompositionRevision
		want   *v1beta1.CompositionRevision
	}{
		"GetLatestRevision": {
			reason: "We should return rev2 as the latest revision.",
			args:   []v1beta1.CompositionRevision{*rev3, *rev1, *rev2},
			want:   rev2,
		},
		"NoControlledRevision": {
			reason: "We should return nil since the revision is not controlled by the comp.",
			args:   []v1beta1.CompositionRevision{*rev3},
			want:   nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := LatestRevision(comp, tc.args)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("LatestRevision(comp, revs): -want, +got:\n%s", diff)
			}
		})
	}
}
