package vcd

import (
	"fmt"
    "log"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	//"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNetworkSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdNsxvSecurityGroupCreate,
		//Read: resourceVcdNsxvSecurityGroupRead,
		//Update: resourceVcdNsxvSecurityGroupUpdate,
		//Delete: resourceVceNsxvSecurityGroupDelete,
		/*Importer: &schema.ResourceImporter{
		    State: resourceVcdNetworkSecurityGroup,
		},*/

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
			},
			"member": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"member_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"member_set": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"member_ids": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"exclude_memeber": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"member_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"exclude_memeber_set": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"member_ids": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"dynamic_member_set": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"operator": {
							Type:     schema.TypeString,
							Required: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"criteria": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"is_valid": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},
		},
	}
}

func resourceVcdNsxvSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
    log.Printf("[DEBUG] Creating security group with name %s", d.Get("name"))
	vcdClient := meta.(*VCDClient)

    _, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}
	// Getting members
	sgMember := d.Get("member").(*schema.Set)
	sgm := append([]*types.SecurityGroupMember{}, sgMemberSchemaToVDC(sgMember, false)...)
    
    // Getting members list
	sgMemberSet := d.Get("member_set").(*schema.Set)
	sgm = append(sgm, sgMemberSchemaToVDC(sgMemberSet, true)...)


	// Create the security group
	sg := types.SecurityGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
        Member: sgm,
	}
	// Adding excluding
	var sgem []*types.SecurityGroupMember
	
    createdSecGroup, err := vdc.CreateNsxvSecurityGroup(&sg)
    if err != nil {
        return fmt.Errorf("error creating new security group: %s", err)
    }

    log.Printf("[DEBUG] Security group with name %s created. Id: %s", createdSecGroup.Name, createdSecGroup.ID)
    d.SetId(createdSecGroup.ID)
	return nil
}

// Convert a list of members from TF schema to VDC
func sgMemberSchemaToVDC(ml *schema.Set, isSet bool) []*types.SecurityGroupMember {
	sgm := []*types.SecurityGroupMember{}

	// Looping over the schemas in the set
	for _, entry := range ml.List() {
		// converting the schema to a map. Both will have a type
		// And the id is either singular or a set
		data := entry.(map[string]interface{})
		sgType := data["type"].(string)
		// Checking if this is a set
		if isSet {
			idsSet := data["member_ids"].([][]string)
			for _, idSet := range idsSet {
				for _, memberId := range idSet {
					sgm = append(sgm, createSgMember(memberId, sgType))
				}
			}
		} else {
			sgm = append(sgm, createSgMember(data["member_id"].(string), sgType))
		}
	}

	return sgm
}

// Converts a schema security group member to a VDC type SecurityGroupMember
func createSgMember(memberId string, member_type string) *types.SecurityGroupMember {
	// Creating the type
	sgt := &types.SecurityGroupType{
		TypeName: member_type,
	}
	// Returning a security group
	return &types.SecurityGroupMember{
		ID: memberId,
		Type:     sgt,
	}
}