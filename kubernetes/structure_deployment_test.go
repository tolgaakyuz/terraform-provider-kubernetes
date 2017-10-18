package kubernetes

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/kubernetes/pkg/api/v1"
	ex_v1beta1 "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
)

// "type": ex_v1beta1.RecreateDeploymentStrategyType,
// "rolling_update": &ex_v1beta1.RollingUpdateDeployment{
// "max_unavailable": 10,
// "max_surge":       20,
// },

func generateDumyLabelSelectorMap() map[string]interface{} {
	return map[string]interface{}{
		"match_labels": map[string]interface{}{
			"a": "v",
			"b": "v2",
		},
		"match_expressions": []interface{}{
			map[string]interface{}{
				"key":      "key",
				"operator": "In",
				"values":   schema.NewSet(schema.HashString, []interface{}{"a", "b"}),
			},
			map[string]interface{}{
				"key":      "key2",
				"operator": "NotIn",
				"values":   schema.NewSet(schema.HashString, []interface{}{"a2", "b2"}),
			},
		},
	}
}

func generateDumyLabelSelector() *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"a": "v",
			"b": "v2",
		},
		MatchExpressions: []metav1.LabelSelectorRequirement{
			metav1.LabelSelectorRequirement{
				Key:      "key",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"a", "b"},
			},
			metav1.LabelSelectorRequirement{
				Key:      "key2",
				Operator: metav1.LabelSelectorOpNotIn,
				Values:   []string{"a2", "b2"},
			},
		},
	}
}

func TestExpandDeploymentSpec(t *testing.T) {
	cases := []struct {
		Input    []interface{}
		Expected ex_v1beta1.DeploymentSpec
	}{
		{
			Input: []interface{}{
				map[string]interface{}{
					"min_ready_seconds":         10,
					"paused":                    false,
					"progress_deadline_seconds": 5,
					"replicas":                  7,
					"revision_history_limit":    6,
					"selector":                  []interface{}{},
					"template":                  []interface{}{},
					"rollback_to":               []interface{}{},
					"strategy":                  []interface{}{},
				},
			},
			Expected: ex_v1beta1.DeploymentSpec{
				Replicas:                ptrToInt32(int32(7)),
				Selector:                &metav1.LabelSelector{},
				Template:                v1.PodTemplateSpec{},
				Strategy:                ex_v1beta1.DeploymentStrategy{},
				MinReadySeconds:         int32(10),
				RevisionHistoryLimit:    ptrToInt32(int32(6)),
				Paused:                  false,
				RollbackTo:              &ex_v1beta1.RollbackConfig{},
				ProgressDeadlineSeconds: ptrToInt32(int32(5)),
			},
		},
	}

	for _, tc := range cases {
		output := expandDeploymentSpec(tc.Input)
		if !reflect.DeepEqual(output, tc.Expected) {
			t.Fatalf("Unexpected output from flattener.\nExpected: %#v\nGiven:    %#v",
				tc.Expected, output)
		}
	}
}

func TestExpandRollbackToReferance(t *testing.T) {
	cases := []struct {
		Input    []interface{}
		Expected *ex_v1beta1.RollbackConfig
	}{
		{
			Input: []interface{}{
				map[string]interface{}{
					"revision": 10,
				},
			},
			Expected: &ex_v1beta1.RollbackConfig{
				Revision: int64(10),
			},
		},
	}

	for _, tc := range cases {
		output := expandRollbackToReferance(tc.Input)
		if !reflect.DeepEqual(output, tc.Expected) {
			t.Fatalf("Unexpected output from flattener.\nExpected: %#v\nGiven:    %#v",
				tc.Expected, output)
		}
	}
}

func TestExpandSelectorReferance(t *testing.T) {
	cases := []struct {
		Input    []interface{}
		Expected *metav1.LabelSelector
	}{
		{
			Input: []interface{}{
				generateDumyLabelSelectorMap(),
			},
			Expected: generateDumyLabelSelector(),
		},
	}

	for _, tc := range cases {
		output := expandSelectorReferance(tc.Input)
		if !reflect.DeepEqual(output, tc.Expected) {
			t.Fatalf("Unexpected output from flattener.\nExpected: %#v\nGiven:    %#v",
				tc.Expected, output)
		}
	}
}
