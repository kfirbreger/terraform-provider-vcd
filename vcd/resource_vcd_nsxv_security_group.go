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
		Read: resourceVcdNsxvSecurityGroupRead,
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
	sgm := append([]*types.SecurityGroupMember{}, sgMemberSchemaToVDC(sgMember)...)
    
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

func genericVcdNsxvSecurityGroupRead(d * schema.ResourceData, meta interface{}, origin string) error {
    log.Printf("[DEBUG] Reading Security Group with ID %s", d.Id())
    vcdClient := meta.(*VCDClient)

    _, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
    if err != nil {
        return fmt.Errorf(errorRetrievingOrgAndVdc, err)
    }

    var secGroup *types.SecurityGroup

    if origin == "datasource" {
        secGroup, err = vdc.GetNsxvSecurityGroupByName(d.Get("name").(string))
    } else {
        secGroup, err = vdc.GetNsxvSecurityGroupById(d.Id())
    }

    if govcd.IsNotFound(err) && origin == "resource" {
        log.Printf("[INFO] unable to find security group with ID %s: %s. Removing from state", d.Id(), err)
        d.SetId("")
        return nil
    }

    if err != nil {
        return fmt.Errorf("unable to find security group with ID %s: %s", d.Id(), err)
    }

    // Persisting to file
    err = setSecurityGroupDate(d, secGroup, vdc, origin)
    if err != nil {
        return fmt.Errorf("unable to store data in statefile: %s", err)
    }

    if origin == "datasource" {
        d.SetId(secGroup.ID)
    }

    log.Printf("[DEBUG] Read security group with ID %s", d.Id())
    return nil
}

func setSecurityGroupData(d *schema.ResourceData, secGroup *types.SecurityGroup, vdc * govcd.Vdc, origin string) error {
    if origin == "resource" {
        d.Set("name", secGroup.Name)
    }

    d.Set("description", secGroup.Description)

    Convert all the members

// Convert a list of members from TF schema to VDC
func sgMemberSchemaToVDC(ml *schema.Set) []*types.SecurityGroupMember {
	sgm := []*types.SecurityGroupMember{}

	// Looping over the schemas in the set
	for _, entry := range ml.List() {
		// converting the schema to a map. Both will have a type
		// And the id is either singular or a set
		data := entry.(map[string]interface{})
        sgm = append(sgm, createSgMember(data["member_id"].(string), data["type"].(string)))
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
