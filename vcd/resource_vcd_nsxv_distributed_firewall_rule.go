package vcd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdDistributedFirewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdDistributedFirewallRuleCreate,
		Read:   resourceVcdDistributedFirewallRuleRead,
		Delete: resourceVcdDistributedFirewallRuleDelete,
		Update: resourceVcdDistributedFirewallRuleUpdate,

		Schema: map[string]*schema.Schema{
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
			// FIXME think about rule ordering (see e.g. resource_vcd_nsxv_firewall_rule.go (before rule_id)
			/*"prepend": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},*/
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
						"group_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

//func getDistributedFirewallRuleData() Terraform -> Govcd types
//func setDistributedFirewallRuleData() Govdc types -> Terraform

func resourceVcdDistributedFirewallRuleCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

/*func resourceVcdDistributedFirewallRuleCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Config)
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	randomName, err := randStringBytes(20)
	if err != nil {
		return errwrap.Wrapf("Unable to generate random string. Reason: {{err}}", err)
	}

	var payload distributedFirewallConfiguration
	etag, err := c.distributedFirewallGetRequest("/network/firewall/globalroot-0/config/layer3sections/"+c.VcdID, &payload)
	if err != nil {
		return errwrap.Wrapf("Fetching firewall rules failed. Reason: {{err}}", err)
	}

	var rule distributedFirewallRule
	rule.Name = d.Get("name").(string) + " - " + randomName
	rule.Disabled = d.Get("disabled").(bool)
	rule.Action = d.Get("action").(string)
	rule.PacketType = "any"
	rule.Direction = d.Get("direction").(string)
	appliedToItem := distributedFirewallItem{
		Name:    "VDC",
		Value:   c.VcdID,
		Type:    "VDC",
		IsValid: true,
	}
	rule.AppliedToList.AppliedTo = append(rule.AppliedToList.AppliedTo, appliedToItem)

	rule.Sources.Excluded = d.Get("source_exclude").(bool)
	var vmName string
	if sources, ok := d.GetOk("source"); ok {
		for _, src := range sources.(*schema.Set).List() {
			data := src.(map[string]interface{})
			newSource := distributedFirewallItem{
				Name:    "Source-item",
				IsValid: true,
			}
			if data["type"] == "ip" {
				newSource.Type = "Ipv4Address"
				newSource.Value = data["value"].(string)
			}
			if data["type"] == "vm" {
				newSource.Type = "VirtualMachine"
				vmName = strings.Replace(data["value"].(string), "vm-", "", 1)
				newSource.Value = "urn:vcloud:vm:" + vmName
			}
			if data["type"] == "sg" {
				newSource.Type = "SecurityGroup"
				newSource.Value = data["value"].(string)
			}
			rule.Sources.Source = append(rule.Sources.Source, newSource)
		}
	}

	rule.Destinations.Excluded = d.Get("destination_exclude").(bool)
	if destinations, ok := d.GetOk("destination"); ok {
		for _, dst := range destinations.(*schema.Set).List() {
			data := dst.(map[string]interface{})
			newDest := distributedFirewallItem{
				Name:    "Destination-item",
				IsValid: true,
			}
			if data["type"] == "ip" {
				newDest.Type = "Ipv4Address"
				newDest.Value = data["value"].(string)
			}
			if data["type"] == "vm" {
				newDest.Type = "VirtualMachine"
				vmName = strings.Replace(data["value"].(string), "vm-", "", 1)
				newDest.Value = "urn:vcloud:vm:" + vmName
			}
			if data["type"] == "sg" {
				newDest.Type = "SecurityGroup"
				newDest.Value = data["value"].(string)
			}
			rule.Destinations.Destination = append(rule.Destinations.Destination, newDest)
		}
	}

	if services, ok := d.GetOk("service"); ok {
		for _, srv := range services.(*schema.Set).List() {
			data := srv.(map[string]interface{})
			if data["group_name"] != "" {
				service, err := c.fetchServiceGroup(data["group_name"].(string))
				if err != nil {
					return errwrap.Wrapf("Create firewall rule failed. Reason: {{err}}", err)
				}
				rule.Services.Service = append(rule.Services.Service, service)
			} else {
				service := distributedFirewallService{
					ProtocolName: data["protocol"].(string),
					Protocol:     "6",
				}
				if strings.ToLower(data["protocol"].(string)) == "icmp" {
					service.Protocol = "1"
				} else {
					if strings.ToLower(data["protocol"].(string)) == "udp" {
						service.Protocol = "17"
					}
					service.DestinationPort = data["destination_port"].(string)
					service.SourcePort = data["source_port"].(string)
				}
				rule.Services.Service = append(rule.Services.Service, service)
			}
		}
	}

	if d.Get("prepend").(bool) {
		// Append every item of the payload.Rule slice to a new slice starting with the new rule.
		payload.Rule = append([]distributedFirewallRule{rule}, payload.Rule...)
	} else {
		payload.Rule = append(payload.Rule, rule)
	}

	if err := c.distributedFirewallPut(etag, payload); err != nil {
		return errwrap.Wrapf("Create firewall rule failed. Reason: {{err}}", err)
	}
	d.SetId(randomName)

	return resourceVcdDistributedFirewallRuleRead(d, meta)
}*/

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

func setDistributedFirewallRuleData(d *schema.ResourceData, dfwRule *types.DistributedFirewallRule, vdc *govcd.Vdc) error {
	_ = d.Set("name", dfwRule.Name)

	dfwSection, _, err := vdc.GetNsxvDistributedFirewallSectionById(dfwRule.SectionId)
	if err != nil {
		return fmt.Errorf("could not retrieve distributed firewall section for distributed firewall rule with id %s: %s", dfwRule.ID, err)
	}

	_ = d.Set("section", dfwSection.Type)
	_ = d.Set("action", dfwRule.Action)
	_ = d.Set("disabled", dfwRule.Disabled)
	_ = d.Set("source_exclude", dfwRule.Sources.Excluded)
	_ = d.Set("destination_exclude", dfwRule.Destinations.Excluded)
	_ = d.Set("direction", dfwRule.Direction)

	var sources []map[string]string
	for _, source := range dfwRule.Sources.Source {
		sources = append(sources, map[string]string{"value": source.Value, "type": source.Type})
	}
	d.Set("sources", sources)

	var destinations []map[string]string
	for _, destination := range dfwRule.Destinations.Destination {
		destinations = append(destinations, map[string]string{"value": destination.Value, "type": destination.Type})
	}
	d.Set("destinations", destinations)

	var services []map[string]string
	for _, service := range dfwRule.Services.DistributedFirewallRuleService {
		if service.ProtocolName != "" {
			services = append(services, map[string]string{"protocol": service.ProtocolName, "destination_port": service.DestinationPort, "source_port": service.SourcePort})
		}
	}
	d.Set("service", services)
	return nil
}

func resourceVcdDistributedFirewallRuleDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Deleteing distributed firewall with ID %s", d.Id())
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

func resourceVcdDistributedFirewallRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

/*func resourceVcdDistributedFirewallRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Config)
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	var payload distributedFirewallConfiguration
	etag, err := c.distributedFirewallGetRequest("/network/firewall/globalroot-0/config/layer3sections/"+c.VcdID, &payload)
	if err != nil {
		return errwrap.Wrapf("Fetching firewall rules failed. Reason: {{err}}", err)
	}

	ruleUpdated := false
	for i, rule := range payload.Rule {
		if len(rule.Name) >= 20 && rule.Name[len(rule.Name)-20:] == d.Id() {
			payload.Rule[i].Name = d.Get("name").(string) + " - " + d.Id()
			payload.Rule[i].Disabled = d.Get("disabled").(bool)
			payload.Rule[i].Action = d.Get("action").(string)
			payload.Rule[i].Sources.Excluded = d.Get("source_exclude").(bool)
			payload.Rule[i].Destinations.Excluded = d.Get("destination_exclude").(bool)

			var vmName string
			payload.Rule[i].Sources.Source = nil
			payload.Rule[i].Destinations.Destination = nil
			payload.Rule[i].Services.Service = nil

			if sources, ok := d.GetOk("source"); ok {
				for _, src := range sources.(*schema.Set).List() {
					data := src.(map[string]interface{})
					newsource := distributedFirewallItem{
						IsValid: true,
					}
					if data["type"].(string) == "ip" {
						newsource.Type = "Ipv4Address"
						newsource.Value = data["value"].(string)
					} else if data["type"].(string) == "vm" {
						newsource.Type = "VirtualMachine"
						vmName = strings.Replace(data["value"].(string), "vm-", "", 1)
						newsource.Value = "urn:vcloud:vm:" + vmName
					} else if data["type"].(string) == "sg" {
						newsource.Type = "SecurityGroup"
						newsource.Value = data["value"].(string)
					} else {
						continue
					}
					payload.Rule[i].Sources.Source = append(payload.Rule[i].Sources.Source, newsource)
				}
			}
			if destinations, ok := d.GetOk("destination"); ok {
				for _, src := range destinations.(*schema.Set).List() {
					data := src.(map[string]interface{})
					newdest := distributedFirewallItem{
						IsValid: true,
					}
					if data["type"].(string) == "ip" {
						newdest.Type = "Ipv4Address"
						newdest.Value = data["value"].(string)
					} else if data["type"].(string) == "vm" {
						newdest.Type = "VirtualMachine"
						vmName = strings.Replace(data["value"].(string), "vm-", "", 1)
						newdest.Value = "urn:vcloud:vm:" + vmName
					} else if data["type"].(string) == "sg" {
						newdest.Type = "SecurityGroup"
						newdest.Value = data["value"].(string)
					} else {
						continue
					}
					payload.Rule[i].Destinations.Destination = append(payload.Rule[i].Destinations.Destination, newdest)
				}
			}

			if services, ok := d.GetOk("service"); ok {
				for _, srv := range services.(*schema.Set).List() {
					data := srv.(map[string]interface{})
					if data["group_name"] != "" {
						service, err := c.fetchServiceGroup(data["group_name"].(string))
						if err != nil {
							return errwrap.Wrapf("Update firewall rule failed. Reason: {{err}}", err)
						}
						payload.Rule[i].Services.Service = append(payload.Rule[i].Services.Service, service)
					} else {

						service := distributedFirewallService{
							ProtocolName: data["protocol"].(string),
							Protocol:     "6",
						}
						if strings.ToLower(data["protocol"].(string)) == "icmp" {
							service.Protocol = "1"
						} else {
							if strings.ToLower(data["protocol"].(string)) == "udp" {
								service.Protocol = "17"
							}
							service.DestinationPort = data["destination_port"].(string)
							service.SourcePort = data["source_port"].(string)
						}
						payload.Rule[i].Services.Service = append(payload.Rule[i].Services.Service, service)
					}

				}
			}

			if d.HasChange("prepend") {
				updatedRule := payload.Rule[i]

				// Create a new slice with all existing elements exept for the changed rule.
				payload.Rule = append(payload.Rule[:i], payload.Rule[i+1:]...)

				if d.Get("prepend").(bool) {
					// Append every item of the payload.Rule slice to a new slice starting with the new rule.
					payload.Rule = append([]distributedFirewallRule{updatedRule}, payload.Rule...)
				} else {
					payload.Rule = append(payload.Rule, updatedRule)
				}
			}

			ruleUpdated = true
			break
		}
	}

	if ruleUpdated == false {
		d.SetId("")
		return errwrap.Wrapf("Unable to find firewall rule, could not update.", nil)
	}

	if err := c.distributedFirewallPut(etag, payload); err != nil {
		return errwrap.Wrapf("Create firewall rule failed. Reason: {{err}}", err)
	}
	return nil
}*/

/*func (c *Config) distributedFirewallGetRequest(uri string, response interface{}) (string, error) {
	request := newRequest()
	request.Operation = "GET"
	if err := request.setURI(uri, c.Endpoint); err != nil {
		return "", fmt.Errorf("Request failed: %s", err)
	}
	request.AuthTokenType = c.TokenType
	request.AuthToken = c.Token
	request.APIVersion = c.Version
	request.Response = response
	if err := request.buildRequest(); err != nil {
		return "", fmt.Errorf("Request failed: %s", err)
	}
	if err := request.doRequest(); err != nil {
		return "", fmt.Errorf("Request failed: %s", err)
	}
	defer request.RawResponse.Body.Close()
	request.parseResponse(request.Response)
	etag := request.RawResponse.Header.Get("Etag")

	return etag, nil
}

func (c *Config) distributedFirewallPut(etag string, payload interface{}) error {
	content, err := prepareXML(payload)
	if err != nil {
		return fmt.Errorf("Unable to create payload: %s", err)
	}
	log.Printf("[DEBUG] firewall put request. %#v", string(content))

	request := newRequest()
	request.Operation = "PUT"
	if err := request.setURI("/network/firewall/globalroot-0/config/layer3sections/"+c.VcdID+"?async=true", c.Endpoint); err != nil {
		return fmt.Errorf("Request failed: %s", err)
	}
	request.AuthTokenType = c.TokenType
	request.AuthToken = c.Token
	request.APIVersion = c.Version
	request.Content = content
	request.ContentType = "application/xml"
	request.Headers["If-Match"] = etag

	if err := request.buildRequest(); err != nil {
		return fmt.Errorf("Request failed: %s", err)
	}
	if err := request.doRequest(); err != nil {
		return fmt.Errorf("Request failed: %s", err)
	}
	defer request.RawResponse.Body.Close()

	return c.extractTask(request)
}

func (c *Config) distributedFirewallDelete(etag string, uri string, retry int) error {
	request := newRequest()

	request.Operation = "DELETE"
	request.APIVersion = c.Version
	request.AuthTokenType = c.TokenType
	request.AuthToken = c.Token
	request.Headers["If-Match"] = etag

	if err := request.setURI(uri, c.Endpoint); err != nil {
		return fmt.Errorf("Request failed: %s", err)
	}
	if err := request.buildRequest(); err != nil {
		return fmt.Errorf("Request failed: %s", err)
	}

	if err := request.doRequest(); err != nil {
		if retry > 0 {
			// If the error might be recoverable, try again.
			if strings.HasSuffix(err.Error(), "cannot be deleted, because it is in use.") {
				time.Sleep(2 * time.Second)
				return c.distributedFirewallDelete(etag, uri, retry-1)
			}
		}
		return fmt.Errorf("Request failed: %s", err)
	}
	request.RawResponse.Body.Close()

	return nil
}

func (c *Config) fetchServiceGroup(groupName string) (distributedFirewallService, error) {
	var service distributedFirewallService
	var resp serviceGroupList
	if err := c.getRequest("/network/services/applicationgroup/scope/"+c.VcdID, &resp); err != nil {
		return service, errwrap.Wrapf("Unable to fetch service group for dfw. Reason: {{err}}", err)
	}
	log.Printf("Application groups found %#v", resp)
	for _, serviceGroup := range resp.ApplicationGroup {
		if serviceGroup.Name == groupName {
			service.Name = serviceGroup.Name
			service.Value = serviceGroup.ID
			service.Type = "ApplicationGroup"
			return service, nil
		}
	}

	return service, fmt.Errorf("Unable to find service group %s", groupName)
}

type distributedFirewallItem struct {
	Name    string `xml:"name"`
	Value   string `xml:"value"`
	Type    string `xml:"type"`
	IsValid bool   `xml:"isValid"`
}

type distributedFirewallService struct {
	IsValid         bool   `xml:"isvalid,omitempty"`
	SourcePort      string `xml:"sourcePort,omitempty"`
	DestinationPort string `xml:"destinationPort,omitempty"`
	Protocol        string `xml:"protocol,omitempty"`
	ProtocolName    string `xml:"protocolName,omitempty"`
	Name            string `xml:"name,omitempty"`
	Value           string `xml:"value,omitempty"`
	Type            string `xml:"type,omitempty"`
}

type distributedFirewallRule struct {
	ID            string `xml:"id,attr,omitempty"`
	Disabled      bool   `xml:"disabled,attr"`
	Logged        bool   `xml:"logged,attr"`
	Name          string `xml:"name"`
	Action        string `xml:"action"`
	AppliedToList struct {
		AppliedTo []distributedFirewallItem `xml:"appliedTo"`
	} `xml:"appliedToList,omitempty"`
	SectionID int `xml:"sectionId,omitempty"`

	Sources struct {
		Excluded bool                      `xml:"excluded,attr"`
		Source   []distributedFirewallItem `xml:"source"`
	} `xml:"sources,omitempty"`
	Destinations struct {
		Excluded    bool                      `xml:"excluded,attr"`
		Destination []distributedFirewallItem `xml:"destination"`
	} `xml:"destinations"`
	Services struct {
		Service []distributedFirewallService `xml:"service"`
	} `xml:"services,omitempty"`
	Direction  string `xml:"direction"`
	PacketType string `xml:"packetType"`
	Tag        string `xml:"tag,omitempty"`
}

type distributedFirewallConfiguration struct {
	XMLName          xml.Name                  `xml:"section"`
	ID               int                       `xml:"id,attr"`
	Name             string                    `xml:"name,attr"`
	GenerationNumber string                    `xml:"generationNumber,attr"`
	Timestamp        string                    `xml:"timestamp,attr"`
	Type             string                    `xml:"type,attr"`
	Rule             []distributedFirewallRule `xml:"rule"`
}

type applicationGroup struct {
	ID   string `xml:"objectId"`
	Name string `xml:"name"`
}

type serviceGroupList struct {
	XMLName          xml.Name           `xml:"list"`
	ApplicationGroup []applicationGroup `xml:"applicationGroup"`
}*/
