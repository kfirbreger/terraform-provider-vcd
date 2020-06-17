package vcd

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxvDistributedFirewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdDistributedFirewallRuleCreate,
		Read:   resourceVcdDistributedFirewallRuleRead,
		Delete: resourceVcdDistributedFirewallRuleDelete,
		Update: resourceVcdDistributedFirewallRuleUpdate,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"section": {
				Type:     schema.TypeString,
				Required: true,
			},
			"action": {
				Type:     schema.TypeString,
				Required: true,
			},
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"source_exclude": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"destination_exclude": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"direction": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "in",
			},
			"sources": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
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
			"destinations": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
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
			// FIXME: rename to plural services
			"service": { //TODO: Add validation that either group_name or protocol has been set.
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_port": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"destination_port": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Optional: true,
						},
						// FIXME application group is currently not supported
						//"group_name": {
						//	Type:     schema.TypeString,
						//	Optional: true,
						//},
					},
				},
			},
		},
	}
}

//resourceVcdDistributedFirewallRuleCreate Create a new distributed firewall rule
func resourceVcdDistributedFirewallRuleCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Creating new distributed firewall")
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	dfwRule, err := getDistributedFirewallRuleData(d, vdc)
	if err != nil {
		return fmt.Errorf("unable to convert terraform resource to vcd resource for distributed firewall rule create: %s", err)
	}

	createdDfwRule, err := vdc.CreateNsxvDistributedFirewallRule(dfwRule)
	if err != nil {
		return fmt.Errorf("error creating new distributed firewall rule: %s", err)
	}

	d.SetId(createdDfwRule.ID)
	return resourceVcdDistributedFirewallRuleRead(d, meta)
}

// resourceVcdDistributedFirewallRuleRead Read distributed firewall state from api and update terraform data
func resourceVcdDistributedFirewallRuleRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Reading distributed firewall with ID %s", d.Id())
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	id := d.Id()
	readFirewallRule, err := vdc.GetNsxvDistributedFirewallRuleById(id)
	if err != nil {
		d.SetId("")
		return fmt.Errorf("unable to find distributed firewall rule with ID %s: %s", d.Id(), err)
	}

	err = setDistributedFirewallRuleData(d, readFirewallRule, vdc)
	if err != nil {
		return err
	}

	return nil
}

// resourceVcdDistributedFirewallRuleUpdate Update distributed firewall state based on terraform data
func resourceVcdDistributedFirewallRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating distributed firewall with ID %s", d.Id())
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	dfwUpdatedRule, err := getDistributedFirewallRuleData(d, vdc)
	if err != nil {
		return fmt.Errorf("unable to convert terraform resource to vcd resource for distributed firewall rule update: %s", err)
	}
	dfwUpdatedRule.ID = d.Id()

	_, err = vdc.UpdateNsxvDistributedFirewallRule(dfwUpdatedRule)
	if err != nil {
		return fmt.Errorf("unable to update distributed firewall rule with ID %s: %s", d.Id(), err)
	}

	return resourceVcdDistributedFirewallRuleRead(d, meta)
}

// resourceVcdDistributedFirewallRuleDelete Delete distributed firewall
func resourceVcdDistributedFirewallRuleDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Deleting distributed firewall with ID %s", d.Id())
	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	id := d.Id()
	err = vdc.DeleteNsxvDistributedFirewallRule(id)
	if err != nil {
		return fmt.Errorf("error deleting distributed firewall rule with id %s: %s", d.Id(), err)
	}

	d.SetId("")
	return nil
}

// getDistributedFirewallRuleData Parse terraform resource data and convert to govcd type
func getDistributedFirewallRuleData(d *schema.ResourceData, vdc *govcd.Vdc) (*types.DistributedFirewallRule, error) {
	dfwRule := new(types.DistributedFirewallRule)
	dfwRule.Name = d.Get("name").(string)

	dfwSectionType := d.Get("section").(string)
	// FIXME make DRY
	switch dfwSectionType {
	case "LAYER2":
		dfwSection, _, err := vdc.GetNsxvDistributedFirewallLayer2Section()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve distributed firewall section %s ", err)
		}
		dfwRule.SectionId = dfwSection.ID
	case "LAYER3":
		dfwSection, _, err := vdc.GetNsxvDistributedFirewallLayer3Section()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve distributed firewall section %s ", err)
		}
		dfwRule.SectionId = dfwSection.ID
	default:
		return nil, fmt.Errorf("unknown distributed firewall section type: %s", dfwSectionType)
	}

	dfwRule.Action = d.Get("action").(string)
	dfwRule.Disabled = d.Get("disabled").(bool)
	dfwRule.PacketType = "any"
	dfwRule.Direction = d.Get("direction").(string)

	if sources, ok := d.GetOk("sources"); ok {
		dfwRule.Sources = new(types.DistributedFirewallRuleSource)
		for _, source := range sources.(*schema.Set).List() {
			data := source.(map[string]interface{})
			newDfwSubject := new(types.DistributedFirewallRuleSubject)
			newDfwSubject.IsValid = true
			newDfwSubject.Type = data["type"].(string)
			newDfwSubject.Value = data["value"].(string)
			dfwRule.Sources.Source = append(dfwRule.Sources.Source, newDfwSubject)
		}
		dfwRule.Sources.Excluded = d.Get("source_exclude").(bool)
	}

	if destinations, ok := d.GetOk("destinations"); ok {
		dfwRule.Destinations = new(types.DistributedFirewallRuleDestination)
		for _, destination := range destinations.(*schema.Set).List() {
			data := destination.(map[string]interface{})
			newDfwSubject := new(types.DistributedFirewallRuleSubject)
			newDfwSubject.IsValid = true
			newDfwSubject.Type = data["type"].(string)
			newDfwSubject.Value = data["value"].(string)
			dfwRule.Destinations.Destination = append(dfwRule.Destinations.Destination, newDfwSubject)
		}
		dfwRule.Destinations.Excluded = d.Get("destination_exclude").(bool)
	}

	if services, ok := d.GetOk("service"); ok {
		dfwRule.Services = new(types.DistributedFirewallRuleServices)
		for _, service := range services.(*schema.Set).List() {
			data := service.(map[string]interface{})
			newDfwService := new(types.DistributedFirewallRuleService)
			protocol := strings.ToLower(data["protocol"].(string))
			switch protocol {
			// TODO lookup protocol ids and add constants. What about any?
			case "icmp":
				newDfwService.Protocol = "1"
				newDfwService.ProtocolName = "ICMP"
			case "tcp":
				newDfwService.Protocol = "6"
				newDfwService.ProtocolName = "TCP"
			case "udp":
				newDfwService.Protocol = "17"
				newDfwService.ProtocolName = "UDP"
			}
			newDfwService.DestinationPort = data["destination_port"].(string)
			newDfwService.SourcePort = data["source_port"].(string)
			dfwRule.Services.DistributedFirewallRuleService = append(dfwRule.Services.DistributedFirewallRuleService, newDfwService)
		}
	}

	return dfwRule, nil
}

// setDistributedFirewallRuleData Convert govcd distributed firewall data to terraform data
func setDistributedFirewallRuleData(d *schema.ResourceData, dfwRule *types.DistributedFirewallRule, vdc *govcd.Vdc) error {
	_ = d.Set("name", dfwRule.Name)

	dfwSection, _, err := vdc.GetNsxvDistributedFirewallSectionById(dfwRule.SectionId)
	if err != nil {
		return fmt.Errorf("could not retrieve distributed firewall section for distributed firewall rule with id %s: %s", dfwRule.ID, err)
	}

	_ = d.Set("section", dfwSection.Type)
	_ = d.Set("action", dfwRule.Action)
	_ = d.Set("disabled", dfwRule.Disabled)
	_ = d.Set("direction", dfwRule.Direction)

	if dfwRule.Sources != nil {
		_ = d.Set("source_exclude", dfwRule.Sources.Excluded)
		var sources []map[string]string
		for _, source := range dfwRule.Sources.Source {
			sources = append(sources, map[string]string{"value": source.Value, "type": source.Type})
		}
		d.Set("sources", sources)
	}

	if dfwRule.Destinations != nil {
		_ = d.Set("destination_exclude", dfwRule.Destinations.Excluded)
		var destinations []map[string]string
		for _, destination := range dfwRule.Destinations.Destination {
			destinations = append(destinations, map[string]string{"value": destination.Value, "type": destination.Type})
		}
		d.Set("destinations", destinations)
	}

	if dfwRule.Services != nil {
		var services []map[string]string
		for _, service := range dfwRule.Services.DistributedFirewallRuleService {
			if service.ProtocolName != "" {
				services = append(services, map[string]string{"protocol": service.ProtocolName, "destination_port": service.DestinationPort, "source_port": service.SourcePort})
			}
		}
		d.Set("service", services)
	}
	return nil
}
