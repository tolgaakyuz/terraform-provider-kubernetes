package kubernetes

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	ex_v1beta1 "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	kubernetes "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

func resourceKubernetesDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubernetesDeploymentCreate,
		Read:   resourceKubernetesDeploymentRead,
		Exists: resourceKubernetesDeploymentExists,
		Update: resourceKubernetesDeploymentUpdate,
		Delete: resourceKubernetesDeploymentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"metadata": namespacedMetadataSchema("deployment", true),
			"spec": {
				Type:        schema.TypeList,
				Description: "Specification of the desired behavior of the Deployment.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_ready_seconds": {
							Type:        schema.TypeInt,
							Description: "Minimum number of seconds for which a newly created pod should be ready without any of its container crashing, for it to be considered available",
							Optional:    true,
						},
						"paused": {
							Type:        schema.TypeBool,
							Description: "Indicates that the deployment is paused.",
							Optional:    true,
						},
						"progress_deadline_seconds ": {
							Type:        schema.TypeInt,
							Description: "The maximum time in seconds for a deployment to make progress before it is considered to be failed. The deployment controller will continue to process failed deployments and a condition with a ProgressDeadlineExceeded reason will be surfaced in the deployment status. Once autoRollback is implemented, the deployment controller will automatically rollback failed deployments. Note that progress will not be estimated during the time a deployment is paused",
							Optional:    true,
							Default:     600,
						},
						"replicas": {
							Type:        schema.TypeInt,
							Description: "Number of desired pods. This is a pointer to distinguish between explicit zero and not specified.",
							Optional:    true,
							Default:     1,
						},
						"revision_history_limit": {
							Type:        schema.TypeInt,
							Description: "The number of old ReplicaSets to retain to allow rollback. This is a pointer to distinguish between explicit zero and not specified.",
							Optional:    true,
							Default:     2,
						},
						"rollback_to": {
							Type:        schema.TypeList,
							Description: "The config this deployment is rolling back to. Will be cleared after rollback is done.",
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: generateRollbackTo(),
							},
						},
						"selector": {
							Type:        schema.TypeList,
							Description: "Label selector for pods. Existing ReplicaSets whose pods are selected by this will be the ones affected by this deployment.",
							Optional:    true,
							ForceNew:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: generateLabelSelector(),
							},
						},
						"strategy": {
							Type:        schema.TypeList,
							Description: "The deployment strategy to use to replace existing pods with new ones.",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"rolling_update": {
										Type:        schema.TypeList,
										Description: "Rolling update config params. Present only if DeploymentStrategyType = RollingUpdate.",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"max_surge": {
													Type:        schema.TypeString,
													Description: "The maximum number of pods that can be scheduled above the desired number of pods. Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%). This can not be 0 if MaxUnavailable is 0. Absolute number is calculated from percentage by rounding up. By default, a value of 1 is used. Example: when this is set to 30%, the new RC can be scaled up immediately when the rolling update starts, such that the total number of old and new pods do not exceed 130% of desired pods. Once old pods have been killed, new RC can be scaled up further, ensuring that total number of pods running at any time during the update is atmost 130% of desired pods.",
													Optional:    true,
												},
												"max_unavailable": {
													Type:        schema.TypeString,
													Description: "The maximum number of pods that can be unavailable during the update. Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%). Absolute number is calculated from percentage by rounding down. This can not be 0 if MaxSurge is 0. By default, a fixed value of 1 is used. Example: when this is set to 30%, the old RC can be scaled down to 70% of desired pods immediately when the rolling update starts. Once new pods are ready, old RC can be scaled down further, followed by scaling up the new RC, ensuring that the total number of pods available at all times during the update is at least 70% of desired pods.",
													Optional:    true,
												},
											},
										},
									},
									"type": {
										Type:        schema.TypeString,
										Description: "Type of deployment. Can be 'Recreate' or 'RollingUpdate'.",
										Optional:    true,
										Default:     "RollingUpdate",
									},
								},
							},
						},
						"template": {
							Type:        schema.TypeList,
							Description: "Template describes the pods that will be created.",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"metadata": namespacedMetadataSchema("template", true),
									"spec": {
										Type:        schema.TypeList,
										Description: "Specification of the desired behavior of the pod.",
										Optional:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: podSpecFields(false),
										},
									},
								},
							},
						},
					},
				},
			},
			"status": {
				Type:        schema.TypeList,
				Description: "Most recently observed status of the Deployment.",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"available_replicas": {
							Type:        schema.TypeInt,
							Description: "Total number of available pods (ready for at least minReadySeconds) targeted by this deployment.",
							Optional:    true,
						},
						"conditions": {
							Type:        schema.TypeList,
							Description: "Represents the latest available observations of a deployment's current state.",
							Optional:    true,
							Elem:        generateDeploymentCondition(),
						},
						"observed_generation": {
							Type:        schema.TypeInt,
							Description: "The generation observed by the deployment controller",
							Optional:    true,
						},
						"ready_replicas": {
							Type:        schema.TypeInt,
							Description: "Total number of ready pods targeted by this deployment.",
							Optional:    true,
						},
						"replicas": {
							Type:        schema.TypeInt,
							Description: "Total number of non-terminated pods targeted by this deployment (their labels match the selector).",
							Optional:    true,
						},
						"unavailable_replicas": {
							Type:        schema.TypeInt,
							Description: "Total number of unavailable pods targeted by this deployment.",
							Optional:    true,
						},
						"updated_replicas": {
							Type:        schema.TypeInt,
							Description: "Total number of non-terminated pods targeted by this deployment that have the desired template spec.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func resourceKubernetesDeploymentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	svc := ex_v1beta1.Deployment{
		ObjectMeta: metadata,
		Spec:       expandDeploymentSpec(d.Get("spec").([]interface{})),
		Status:     expandDeploymentStatus(d.Get("status").([]interface{})),
	}
	log.Printf("[INFO] Creating new deployment: %#v", svc)
	out, err := conn.ExtensionsV1beta1().Deployments(metadata.Namespace).Create(&svc)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Submitted new deployment: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceKubernetesDeploymentRead(d, meta)
}

func resourceKubernetesDeploymentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}
	log.Printf("[INFO] Reading deployment %s", name)
	svc, err := conn.ExtensionsV1beta1().Deployments(namespace).Get(name, v1.GetOptions{})
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received deployment: %#v", svc)
	err = d.Set("metadata", flattenMetadata(svc.ObjectMeta))
	if err != nil {
		return err
	}

	flattenedSpec := flattenDeploymentSpec(svc.Spec)
	log.Printf("[DEBUG] Flattened deployment spec: %#v", flattenedSpec)
	err = d.Set("spec", flattenedSpec)
	if err != nil {
		return err
	}

	flattenedStatus := flattenDeploymentStatus(svc.Status)
	log.Printf("[DEBUG] Flattened deployment status: %#v", flattenedStatus)
	err = d.Set("status", flattenedStatus)
	if err != nil {
		return err
	}

	return nil
}

func resourceKubernetesDeploymentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	ops := patchMetadata("metadata.0.", "/metadata/", d)
	if d.HasChange("spec") {
		diffOps := patchDeploymentSpec("spec.0.", "/spec", d)
		ops = append(ops, diffOps...)
	}
	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}
	log.Printf("[INFO] Updating deployment %q: %v", name, string(data))
	out, err := conn.ExtensionsV1beta1().Deployments(namespace).Patch(name, types.JSONPatchType, data)
	if err != nil {
		return fmt.Errorf("Failed to update deployment: %s", err)
	}
	log.Printf("[INFO] Submitted updated deployment: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceKubernetesHorizontalPodAutoscalerRead(d, meta)
}

func resourceKubernetesDeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*kubernetes.Clientset)

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}
	log.Printf("[INFO] Deleting deployment: %#v", name)
	err = conn.ExtensionsV1beta1().Deployments(namespace).Delete(name, &v1.DeleteOptions{})
	if err != nil {
		return err
	}

	log.Printf("[INFO] Deployment %s deleted", name)

	d.SetId("")
	return nil
}

func resourceKubernetesDeploymentExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	conn := meta.(*kubernetes.Clientset)

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return false, err
	}

	log.Printf("[INFO] Checking deployment %s", name)
	_, err = conn.ExtensionsV1beta1().Deployments(namespace).Get(name, v1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	return true, err
}

func generateRollbackTo() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"revision": {
			Type:        schema.TypeInt,
			Description: "The revision to rollback to. If set to 0, rollback to the last revision.",
			Optional:    true,
		},
	}
}

func generateLabelSelector() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"match_expressions": {
			Type:        schema.TypeList,
			Description: "A list of label selector requirements. The requirements are ANDed.",
			Optional:    true,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"key": {
						Type:        schema.TypeString,
						Description: "The label key that the selector applies to.",
						Optional:    true,
						ForceNew:    true,
					},
					"operator": {
						Type:        schema.TypeString,
						Description: "A key's relationship to a set of values. Valid operators ard `In`, `NotIn`, `Exists` and `DoesNotExist`.",
						Optional:    true,
						ForceNew:    true,
					},
					"values": {
						Type:        schema.TypeSet,
						Description: "An array of string values. If the operator is `In` or `NotIn`, the values array must be non-empty. If the operator is `Exists` or `DoesNotExist`, the values array must be empty. This array is replaced during a strategic merge patch.",
						Optional:    true,
						ForceNew:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Set:         schema.HashString,
					},
				},
			},
		},
		"match_labels": {
			Type:        schema.TypeMap,
			Description: "A map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of `match_expressions`, whose key field is \"key\", the operator is \"In\", and the values array contains only \"value\". The requirements are ANDed.",
			Optional:    true,
			ForceNew:    true,
		},
	}
}

func generateDeploymentCondition() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"message": {
				Type:        schema.TypeString,
				Description: "A human readable message indicating details about the transition.",
				Optional:    true,
			},
			"reason": {
				Type:        schema.TypeString,
				Description: "The reason for the condition's last transition.",
				Optional:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "Status of the condition, one of True, False, Unknown.",
				Optional:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Type of deployment condition.",
				Optional:    true,
			},
		},
	}
}
