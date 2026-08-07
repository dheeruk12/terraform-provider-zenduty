package zenduty

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Zenduty/zenduty-go-sdk/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceSLA() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateSLA,
		UpdateContext: resourceUpdateSLA,
		DeleteContext: resourceDeleteSLA,
		ReadContext:   wrapReadWith404(resourceReadSLA),
		Importer: &schema.ResourceImporter{
			State: resourceSLAImporter,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"escalations": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unique_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"time": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(-432000, 432000),
						},
						"type": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 2),
						},
						"responders": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"user": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"acknowledge_time": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 432000),
			},
			"resolve_time": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 432000),
			},
			"is_active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func CreateSLA(Ctx context.Context, d *schema.ResourceData, m interface{}) (*client.SLAObj, diag.Diagnostics) {
	newSLA := &client.SLAObj{}
	escalations := d.Get("escalations").([]interface{})
	if v, ok := d.GetOk("name"); ok {
		newSLA.Name = v.(string)
	}
	if v, ok := d.GetOk("description"); ok {
		newSLA.Description = v.(string)
		if emptyString(newSLA.Description) {
			return nil, diag.FromErr(errors.New("description is empty"))
		}
	}
	if v, ok := d.GetOk("acknowledge_time"); ok {
		newSLA.AcknowledgeTime = v.(int)
	}
	if v, ok := d.GetOk("resolve_time"); ok {
		newSLA.ResolveTime = v.(int)
	}
	newSLA.IsActive = d.Get("is_active").(bool)

	newSLA.Escalations = make([]client.SLAEscalations, len(escalations))

	for i, escalation := range escalations {
		escalationMap := escalation.(map[string]interface{})
		newEscalation := client.SLAEscalations{}

		if v, ok := escalationMap["time"]; ok {
			newEscalation.Time = v.(int)

		}
		if v, ok := escalationMap["type"]; ok {
			newEscalation.Type = v.(int)
		}
		if v, ok := escalationMap["unique_id"]; ok {
			newEscalation.UniqueID = v.(string)
			if emptyString(newEscalation.UniqueID) {
				newEscalation.UniqueID = genrateUUID()
			}
		}

		if v, ok := escalationMap["responders"]; ok {
			responderusers := v.([]interface{})
			newEscalation.Responders = make([]client.ResponderUser, len(responderusers))
			for j, responder := range responderusers {
				resonderMap := responder.(map[string]interface{})
				responderuser := client.ResponderUser{}
				if v, ok := resonderMap["user"]; ok {
					responderuser.User = v.(string)
				}
				newEscalation.Responders[j] = responderuser
			}
		}
		newSLA.Escalations[i] = newEscalation
	}
	return newSLA, nil

}

func flattenEscalation(escalations []client.SLAEscalations) []map[string]interface{} {
	result := make([]map[string]interface{}, len(escalations))
	for i, escalation := range escalations {
		result[i] = map[string]interface{}{
			"time":       escalation.Time,
			"type":       escalation.Type,
			"unique_id":  escalation.UniqueID,
			"responders": flattenResponderUser(escalation.Responders),
		}
	}
	return result
}

func flattenResponderUser(responders []client.ResponderUser) []map[string]interface{} {
	result := make([]map[string]interface{}, len(responders))
	for i, responder := range responders {
		result[i] = map[string]interface{}{
			"user": responder.User,
		}
	}
	return result
}

func resourceCreateSLA(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	var teamID string
	if v, ok := d.GetOk("team_id"); ok {
		if emptyString(v.(string)) {
			return diag.FromErr(errors.New("team_id must not be empty"))
		}
		teamID = v.(string)
	}

	var diags diag.Diagnostics
	newSLA, createErr := CreateSLA(Ctx, d, m)
	if createErr != nil {
		return createErr
	}

	sla, err := apiclient.Sla.CreateSLA(teamID, newSLA)

	if err != nil {
		return diag.FromErr(err)
	}

	for i, escalation := range sla.Escalations {
		log.Printf("SLAEscalation %d: %v", i, escalation)
	}
	log.Printf("%+v\n", sla.Escalations)
	d.SetId(sla.UniqueID)
	if err := d.Set("escalations", flattenEscalation(sla.Escalations)); err != nil {
		return diag.FromErr(err)
	}

	return diags

}

func resourceUpdateSLA(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	var teamID string
	if v, ok := d.GetOk("team_id"); ok {
		if emptyString(v.(string)) {
			return diag.FromErr(errors.New("team_id must not be empty"))
		}
		teamID = v.(string)
	}

	newSLA, createErr := CreateSLA(Ctx, d, m)
	if createErr != nil {
		return createErr
	}
	id := d.Id()
	var diags diag.Diagnostics

	sla, err := apiclient.Sla.UpdateSLAByID(teamID, id, newSLA)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("escalations", flattenEscalation(sla.Escalations)); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceDeleteSLA(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	var teamID string
	if v, ok := d.GetOk("team_id"); ok {
		if emptyString(v.(string)) {
			return diag.FromErr(errors.New("team_id must not be empty"))
		}
		teamID = v.(string)
	}

	id := d.Id()

	var diags diag.Diagnostics
	err := apiclient.Sla.DeleteSLAByID(teamID, id)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceReadSLA(Ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()

	teamID := d.Get("team_id").(string)
	if v, ok := d.GetOk("team_id"); ok {
		teamID = v.(string)
	}

	id := d.Id()
	if teamID == "" {
		return diag.FromErr(errors.New("team_id is required "))
	}
	var diags diag.Diagnostics
	sla, err := apiclient.Sla.GetSLAByID(teamID, id)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("name", sla.Name)
	d.Set("description", sla.Description)
	d.Set("acknowledge_time", sla.AcknowledgeTime)
	d.Set("resolve_time", sla.ResolveTime)
	d.Set("is_active", sla.IsActive)
	if err := d.Set("escalations", flattenEscalation(sla.Escalations)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceSLAImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected format of id (%q), expected <team_id>/<sla_id>", d.Id())
	} else if !IsValidUUID(parts[0]) {
		return nil, fmt.Errorf("invalid team_id (%q)", parts[0])
	} else if !IsValidUUID(parts[1]) {
		return nil, fmt.Errorf("invalid sla id (%q)", parts[1])
	}
	d.Set("team_id", parts[0])
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}
