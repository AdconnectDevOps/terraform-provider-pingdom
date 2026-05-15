package pingdom

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePingdomTeam() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePingdomTeamRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"member_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
		},
	}
}

func dataSourcePingdomTeamRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	name := d.Get("name").(string)
	teams, err := client.Teams.List()
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error retrieving team: %s", err))
	}
	var found *TeamResponse
	for _, team := range teams {
		if team.Name == name {
			log.Printf("Team: %v", team)
			found = &team
			break
		}
	}
	if found == nil {
		return diag.FromErr(fmt.Errorf("Team '%s' not found", name))
	}
	if err = d.Set("name", found.Name); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting name: %s", err))
	}

	var memberIds []int
	for _, member := range found.Members {
		memberIds = append(memberIds, member.ID)
	}

	if err = d.Set("member_ids", memberIds); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d", found.ID))
	return nil
}
