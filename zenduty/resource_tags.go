package zenduty

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Zenduty/zenduty-go-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTags() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateTags,
		UpdateContext: resourceUpdateTags,
		DeleteContext: resourceDeleteTags,
		ReadContext:   wrapReadWith404(resourceReadTag),
		Importer: &schema.ResourceImporter{
			State: resourceTagImporter,
		},
		Schema: map[string]*schema.Schema{

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"color": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"team_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateUUID(),
			},
		},
	}
}

func validateTags(Ctx context.Context, d *schema.ResourceData, m interface{}) (*client.Tag, diag.Diagnostics) {
	name := d.Get("name").(string)
	color := d.Get("color").(string)
	team := d.Get("team_id").(string)
	newTag := &client.Tag{}
	if !IsValidUUID(team) {
		return nil, diag.FromErr(errors.New("team_id must be a valid UUID"))
	}
	if color != "" && !checkList(color, []string{"magenta", "red", "volcano", "orange", "gold", "lime", "green", "cyan", "blue", "geekblue", "purple"}) {
		return nil, diag.FromErr(errors.New("color must be one of the following: magenta, red, volcano, orange, gold, lime, green, cyan, blue, geekblue, purple"))
	}

	newTag.Name = name
	newTag.Color = color

	return newTag, nil
}

func resourceCreateTags(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	team := d.Get("team_id").(string)
	apiclient, _ := m.(*Config).Client()
	newtag, validationerr := validateTags(ctx, d, m)
	if validationerr != nil {
		return validationerr
	}
	tag, err := apiclient.Tags.CreateTag(team, newtag)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tag.UniqueID)
	return nil
}

func resourceUpdateTags(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	team := d.Get("team_id").(string)
	apiclient, _ := m.(*Config).Client()
	newtag, validationerr := validateTags(ctx, d, m)
	if validationerr != nil {
		return validationerr
	}
	tag, err := apiclient.Tags.UpdateTag(team, d.Id(), newtag)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tag.UniqueID)
	return nil
}

func resourceDeleteTags(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	team := d.Get("team_id").(string)
	apiclient, _ := m.(*Config).Client()

	err := apiclient.Tags.DeleteTag(team, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceReadTag(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiclient, _ := m.(*Config).Client()
	team := d.Get("team_id").(string)
	tag, err := apiclient.Tags.GetTagID(team, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tag.UniqueID)
	d.Set("name", tag.Name)
	d.Set("color", tag.Color)
	d.Set("team_id", tag.Team)
	return nil
}

func resourceTagImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected format of id (%q), expected <team_id>/<tag_id>", d.Id())
	} else if !IsValidUUID(parts[0]) {
		return nil, fmt.Errorf("invalid team_id (%q)", parts[0])
	} else if !IsValidUUID(parts[1]) {
		return nil, fmt.Errorf("invalid tag_id (%q)", parts[1])
	}
	d.Set("team_id", parts[0])
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}
