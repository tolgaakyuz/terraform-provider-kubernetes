package kubernetes

import (
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	v1 "k8s.io/kubernetes/pkg/api/v1"
	ex_v1beta1 "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
)

// Expanders

func expandDeploymentSpec(in []interface{}) ex_v1beta1.DeploymentSpec {
	if len(in) == 0 || in[0] == nil {
		return ex_v1beta1.DeploymentSpec{}
	}

	spec := ex_v1beta1.DeploymentSpec{}
	m := in[0].(map[string]interface{})
	if v, ok := m["min_ready_seconds"]; ok {
		spec.MinReadySeconds = int32(v.(int))
	}
	if v, ok := m["paused"]; ok {
		spec.Paused = v.(bool)
	}
	if v, ok := m["progress_deadline_seconds"].(int); ok {
		spec.ProgressDeadlineSeconds = ptrToInt32(int32(v))
	}
	if v, ok := m["replicas"].(int); ok && v > 0 {
		spec.Replicas = ptrToInt32(int32(v))
	}
	if v, ok := m["revision_history_limit"].(int); ok && v > 0 {
		spec.RevisionHistoryLimit = ptrToInt32(int32(v))
	}
	if v, ok := m["selector"]; ok {
		spec.Selector = expandSelectorReferance(v.([]interface{}))
	}
	if v, ok := m["rollback_to"]; ok {
		spec.RollbackTo = expandRollbackToReferance(v.([]interface{}))
	}
	if v, ok := m["strategy"]; ok {
		spec.Strategy = expandStrategyReferance(v.([]interface{}))
	}
	return spec
}

func expandRollbackToReferance(in []interface{}) *ex_v1beta1.RollbackConfig {
	if len(in) == 0 || in[0] == nil {
		return &ex_v1beta1.RollbackConfig{}
	}

	rollbackConfig := ex_v1beta1.RollbackConfig{}
	m := in[0].(map[string]interface{})
	if v, ok := m["revision"]; ok {
		rollbackConfig.Revision = int64(v.(int))
	}
	return &rollbackConfig
}

func expandSelectorReferance(l []interface{}) *metav1.LabelSelector {
	if len(l) == 0 || l[0] == nil {
		return &metav1.LabelSelector{}
	}
	in := l[0].(map[string]interface{})
	obj := &metav1.LabelSelector{}
	if v, ok := in["match_labels"].(map[string]interface{}); ok && len(v) > 0 {
		obj.MatchLabels = expandStringMap(v)
	}
	if v, ok := in["match_expressions"].([]interface{}); ok && len(v) > 0 {
		obj.MatchExpressions = expandLabelSelectorRequirement(v)
	}
	return obj
}

func expandStrategyReferance(in []interface{}) ex_v1beta1.DeploymentStrategy {
	if len(in) == 0 || in[0] == nil {
		return ex_v1beta1.DeploymentStrategy{}
	}

	strategy := ex_v1beta1.DeploymentStrategy{}
	m := in[0].(map[string]interface{})
	if v, ok := m["type"].(string); ok {
		strategy.Type = ex_v1beta1.DeploymentStrategyType(v)
	}
	if v, ok := m["rolling_update"]; ok {
		strategy.RollingUpdate = expandRollingUpdateReferance(v.([]interface{}))
	}
	return strategy
}

func expandRollingUpdateReferance(in []interface{}) *ex_v1beta1.RollingUpdateDeployment {
	if len(in) == 0 || in[0] == nil {
		return &ex_v1beta1.RollingUpdateDeployment{}
	}

	rollingUpdate := ex_v1beta1.RollingUpdateDeployment{}
	m := in[0].(map[string]interface{})
	if v, ok := m["max_unavailable"].(string); ok && len(v) > 0 {
		i, err := strconv.Atoi(v)

		if err != nil {
			rollingUpdate.MaxUnavailable = &intstr.IntOrString{
				Type:   intstr.String,
				StrVal: v,
			}
		} else {
			rollingUpdate.MaxUnavailable = &intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: int32(i),
			}
		}
	}
	if v, ok := m["max_surge"].(string); ok && len(v) > 0 {
		i, err := strconv.Atoi(v)

		if err != nil {
			rollingUpdate.MaxSurge = &intstr.IntOrString{
				Type:   intstr.String,
				StrVal: v,
			}
		} else {
			rollingUpdate.MaxSurge = &intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: int32(i),
			}
		}
	}

	return &rollingUpdate
}

func expandTemplateReferance(in []interface{}) (v1.PodTemplateSpec, error) {
	if len(in) == 0 || in[0] == nil {
		return v1.PodTemplateSpec{}, nil
	}

	pt := v1.PodTemplateSpec{}
	m := in[0].(map[string]interface{})
	if v, ok := m["metadata"]; ok {
		pt.ObjectMeta = expandMetadata(v.([]interface{}))
	}
	if v, ok := m["spec"]; ok {
		spec, err := expandPodSpec(v.([]interface{}))
		if err != nil {
			return v1.PodTemplateSpec{}, err
		}
		pt.Spec = spec
	}
	return pt, nil
}

func expandDeploymentStatus(in []interface{}) ex_v1beta1.DeploymentStatus {
	if len(in) == 0 || in[0] == nil {
		return ex_v1beta1.DeploymentStatus{}
	}

	status := ex_v1beta1.DeploymentStatus{}
	m := in[0].(map[string]interface{})
	if v, ok := m["available_replicas"]; ok {
		status.AvailableReplicas = int32(v.(int))
	}
	if v, ok := m["observed_generation"]; ok {
		status.ObservedGeneration = int64(v.(int))
	}
	if v, ok := m["replicas"]; ok {
		status.Replicas = int32(v.(int))
	}
	if v, ok := m["updated_replicas"]; ok {
		status.UpdatedReplicas = int32(v.(int))
	}
	if v, ok := m["ready_replicas"]; ok {
		status.ReadyReplicas = int32(v.(int))
	}
	if v, ok := m["conditions"].([]interface{}); ok && len(v) > 0 {
		conditions, _ := expandDeploymentStatusCondition(v)
		status.Conditions = conditions
	}

	return status
}

func expandDeploymentStatusCondition(in []interface{}) ([]ex_v1beta1.DeploymentCondition, error) {
	if len(in) == 0 {
		return []ex_v1beta1.DeploymentCondition{}, nil
	}

	cnds := make([]ex_v1beta1.DeploymentCondition, len(in))
	for i, c := range in {
		m := c.(map[string]interface{})

		if v, ok := m["type"]; ok {
			cnds[i].Type = ex_v1beta1.DeploymentConditionType(v.(string))
		}
		if v, ok := m["status"]; ok {
			cnds[i].Status = v1.ConditionStatus(v.(string))
		}
		if v, ok := m["reason"]; ok {
			cnds[i].Reason = v.(string)
		}
		if v, ok := m["message"]; ok {
			cnds[i].Message = v.(string)
		}
	}
	return cnds, nil
}

// Flatteners

func flattenDeploymentSpec(spec ex_v1beta1.DeploymentSpec) []interface{} {
	m := make(map[string]interface{}, 0)
	if spec.Replicas != nil {
		m["replicas"] = *spec.Replicas
	}
	m["selector"] = flattenLabelSelectorReferance(spec.Selector)
	m["template"] = flattenTemplateReferance(spec.Template)
	m["strategy"] = flattenStrategyReferance(spec.Strategy)
	m["min_ready_seconds"] = spec.MinReadySeconds
	if spec.RevisionHistoryLimit != nil {
		m["revision_history_limit"] = *spec.RevisionHistoryLimit
	}
	m["paused"] = spec.Paused
	m["rollback_to"] = flattenRollbackToReferance(spec.RollbackTo)
	if spec.ProgressDeadlineSeconds != nil {
		m["progress_deadline_seconds"] = *spec.ProgressDeadlineSeconds
	}
	return []interface{}{m}
}

func flattenLabelSelectorReferance(selector *metav1.LabelSelector) []interface{} {
	m := make(map[string]interface{}, 0)
	m["match_labels"] = selector.MatchLabels
	m["match_expressions"] = flattenLabelSelectorRequirementList(selector.MatchExpressions)
	return []interface{}{m}
}

func flattenTemplateReferance(template v1.PodTemplateSpec) []interface{} {
	m := make(map[string]interface{}, 0)
	m["metadata"] = flattenMetadata(template.ObjectMeta)
	podSpec, _ := flattenPodSpec(template.Spec)
	m["spec"] = podSpec
	return []interface{}{m}
}

func flattenStrategyReferance(in ex_v1beta1.DeploymentStrategy) []interface{} {
	m := make(map[string]interface{}, 0)
	m["type"] = string(in.Type)
	m["rolling_update"] = flattenRollingUpdateDeployment(in.RollingUpdate)
	return []interface{}{m}
}

func flattenRollingUpdateDeployment(in *ex_v1beta1.RollingUpdateDeployment) []interface{} {
	m := make(map[string]interface{}, 0)
	m["max_unavailable"] = flattenIntOrString(*in.MaxUnavailable)
	m["max_surge"] = flattenIntOrString(*in.MaxSurge)
	return []interface{}{m}
}

func flattenRollbackToReferance(in *ex_v1beta1.RollbackConfig) []interface{} {
	m := make(map[string]interface{}, 0)
	m["revision"] = in.Revision
	return []interface{}{m}
}

func flattenDeploymentStatus(in ex_v1beta1.DeploymentStatus) []interface{} {
	m := make(map[string]interface{}, 0)
	m["observed_generation"] = in.ObservedGeneration
	m["replicas"] = in.Replicas
	m["update_replicas"] = in.UpdatedReplicas
	m["ready_replicas"] = in.ReadyReplicas
	m["available_replicas"] = in.AvailableReplicas
	m["unavailable_replicas"] = in.UnavailableReplicas
	conditions := make([]map[string]string, 0)
	for _, val := range in.Conditions {
		conditions = append(conditions, flattenDeploymentConditionReferance(val))
	}
	m["conditions"] = conditions
	return []interface{}{m}
}

func flattenDeploymentConditionReferance(in ex_v1beta1.DeploymentCondition) map[string]string {
	m := make(map[string]string, 0)
	m["type"] = string(in.Type)
	m["status"] = string(in.Status)
	m["reason"] = in.Reason
	m["message"] = in.Message
	return m
}

// Patchers

func patchDeploymentSpec(prefix string, pathPrefix string, d *schema.ResourceData) []PatchOperation {
	ops := make([]PatchOperation, 0)

	if d.HasChange(prefix + "min_ready_seconds") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "/minReadySeconds",
			Value: d.Get(prefix + "min_ready_seconds").(int),
		})
	}
	if d.HasChange(prefix + "paused") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "/paused",
			Value: d.Get(prefix + "paused").(bool),
		})
	}
	if d.HasChange(prefix + "progress_deadline_seconds") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "/progressDeadlineSeconds",
			Value: d.Get(prefix + "progress_deadline_seconds").(int),
		})
	}
	if d.HasChange(prefix + "replicas") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "/replicas",
			Value: d.Get(prefix + "replicas").(int),
		})
	}
	if d.HasChange(prefix + "revision_history_limit") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "/revisionHistoryLimit",
			Value: d.Get(prefix + "revision_history_limit").(int),
		})
	}
	if d.HasChange(prefix + "rollback_to") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "/rollbackTo",
			Value: expandRollbackToReferance(d.Get(prefix + "rollback_to").([]interface{})),
		})
	}
	if d.HasChange(prefix + "selector") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "/selector",
			Value: expandSelectorReferance(d.Get(prefix + "selector").([]interface{})),
		})
	}
	if d.HasChange(prefix + "strategy") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "/strategy",
			Value: expandStrategyReferance(d.Get(prefix + "strategy").([]interface{})),
		})
	}
	if d.HasChange(prefix + "template") {
		value, _ := expandTemplateReferance(d.Get(prefix + "template").([]interface{}))
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "/template",
			Value: value,
		})
	}
	return ops
}
