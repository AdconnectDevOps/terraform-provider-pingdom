package pingdom

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/russellcardullo/go-pingdom/pingdom"
)

func dataSourcePingdomContact() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePingdomContactRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"paused": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"teams": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"sms_notification": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"number": {
							Type:     schema.TypeString,
							Required: true,
						},
						"country_code": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "1",
						},
						"severity": {
							Type:     schema.TypeString,
							Required: true,
						},
						"provider": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "nexmo",
						},
					},
				},
			},
			"email_notification": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"severity": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func dataSourcePingdomContactRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*pingdom.Client)
	name := d.Get("name").(string)
	contacts, err := client.Contacts.List()
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error retrieving contact: %s", err))
	}
	var found *pingdom.Contact
	for _, contact := range contacts {
		if contact.Name == name {
			log.Printf("Contact: %v", contact)
			found = &contact
			break
		}
	}
	if found == nil {
		return diag.FromErr(fmt.Errorf("Contact '%s' not found", name))
	}

	if err = d.Set("name", found.Name); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting name: %s", err))
	}

	teams := []map[string]interface{}{}
	for _, team := range found.Teams {
		teams = append(teams, map[string]interface{}{
			"id":   team.ID,
			"name": team.Name,
		})
	}
	if err = d.Set("teams", teams); err != nil {
		return diag.FromErr(err)
	}

	if err = updateResourceFromContactResponse(d, found); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d", found.ID))
	return nil
}
