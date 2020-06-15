package vcd

import (
	"fmt"
    "log"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNetworkSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdNsxvSecurityGroupCreate,
		Read: resourceVcdNsxvSecurityGroupRead,
		Update: resourceVcdNsxvSecurityGroupUpdate,
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
	sgm := append([]*types.SecurityGroupMember{}, expandSecurityGroupMembers(sgMember)...)
    
	// Create the security group
	sg := types.SecurityGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
        Member: sgm,
	}
	// Adding excluding
    sgExcludeMember := d.Get("exclude_member").(*schema.Set)
	sgem := append([]*types.SecurityGroupMember{}, expandSecurityGroupMembers(sgExcludeMember)...)

    createdSecGroup, err := vdc.CreateNsxvSecurityGroup(&sg)
    if err != nil {
        return fmt.Errorf("error creating new security group: %s", err)
    }

    log.Printf("[DEBUG] Security group with name %s created. Id: %s", createdSecGroup.Name, createdSecGroup.ID)
    d.SetId(createdSecGroup.ID)
	return nil
}

func resourceVcdNsxvSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {
    log.Printf("[DEBUG] updating security group with ID %s", d.Id())

    vcdClient := meta.(*VCDClient)

    _, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
    if err != nil {
        return fmt.Errorf(errorRetrievingOrgAndVdc, err)
    }

    secGroup, err := getSecGroup(d, vdc)
    if err != nil {
        return fmt.Errorf("upable to make security group query: %s", err)
    }
    secGroup.ID = d.Id()

    // updating
    // first the sec group info and then the membership information
    _, err = vdc.UpdateNsxvSecurityGroupInfo(secGroup)
    if err != nil {
        return fmt.Errorf("error updating info of security group with ID %s: %s", d.Id(), err)
    } else {
        log.Printf("[DEBUG] updated security group info. Now updating membership for security group with id: %s", d.Id())
    }
    if d.HasChange("member") || d.HasChange("exclude_member") {
        _, err = vdc.UpdateNsxvSecurityGroupMembership(secGroup)
        if err != nil {
            return fmt.Errorf("error updating memberships for security group with ID %s: %s", d.Id(), err)
        }
    }

    log.Printf("[DEBUG] updated security group with ID %s", d.Id())
    return resourceVcdNsxvSecurityGroupRead(d, meta)
}

func datasourceNsxvSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
    return genericVcdNsxvSecurityGroupRead(d, meta, "datasource")
}

func resourceVcdNsxvSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
    return genericVcdNsxvSecurityGroupRead(d, meta, "resource")
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
    err = setSecurityGroupData(d, secGroup, vdc, origin)
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
    var err error
    if origin == "resource" {
        if err = d.Set("name", secGroup.Name); err != nil {
            return fmt.Errorf("[ERROR] failed setting %s as name for security group: %s", secGroup.Name, err)
        }
    }

    if err = d.Set("description", secGroup.Description); err != nil {
        return fmt.Errorf("[ERROR] failed setting description for security group: %s", err)
    }

    if err = d.Set("member", flattenMembersSet(secGroup.Member)); err != nil {
        return fmt.Errorf("[ERROR] failed to set members set: %s", err)
    }

    if err = d.Set("exclude_member", flattenMembersSet(secGroup.ExcludeMember)); err != nil {
        return fmt.Errorf("[ERROR] failed to set exclude members set: %s", err)
    }
    return nil
}

func flattenMembersSet(secGroupMemberList []*types.SecurityGroupMember) []*map[string]interface{} {
    // Creating the slice
    sgMemberSlice := []*map[string]interface{}{}
    for _, sgMember := range secGroupMemberList {
        member := &map[string]interface{}{
            "Name": sgMember.ID,
            "Type": sgMember.Type.TypeName,
        }
        sgMemberSlice = append(sgMemberSlice, member)
    }
    return sgMemberSlice
}

// Convert a list of members from TF schema to VDC
func expandSecurityGroupMembers(ml *schema.Set) []*types.SecurityGroupMember {
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
