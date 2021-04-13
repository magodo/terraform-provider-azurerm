package firewall

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/arm/network/2020-07-01/armnetwork"
	"github.com/davecgh/go-spew/spew"
	"log"
	"reflect"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	azValidate "github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/locks"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/firewall/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/firewall/validate"
	azSchema "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceFirewallPolicyRuleCollectionGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirewallPolicyRuleCollectionGroupCreateUpdate,
		Read:   resourceFirewallPolicyRuleCollectionGroupRead,
		Update: resourceFirewallPolicyRuleCollectionGroupCreateUpdate,
		Delete: resourceFirewallPolicyRuleCollectionGroupDelete,

		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := parse.FirewallPolicyRuleCollectionGroupID(id)
			return err
		}),

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.FirewallPolicyRuleCollectionGroupName(),
			},

			"firewall_policy_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.FirewallPolicyID,
			},

			"priority": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(100, 65000),
			},

			"application_rule_collection": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"priority": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(100, 65000),
						},
						"action": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(armnetwork.FirewallPolicyFilterRuleCollectionActionTypeAllow),
								string(armnetwork.FirewallPolicyFilterRuleCollectionActionTypeDeny),
							}, false),
						},
						"rule": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validate.FirewallPolicyRuleName(),
									},
									"protocols": {
										Type:     schema.TypeSet,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.StringInSlice([]string{
														string(armnetwork.FirewallPolicyRuleApplicationProtocolTypeHTTP),
														string(armnetwork.FirewallPolicyRuleApplicationProtocolTypeHTTPS),
													}, false),
												},
												"port": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(0, 64000),
												},
											},
										},
									},
									"source_addresses": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
											ValidateFunc: validation.Any(
												validation.IsIPAddress,
												validation.IsCIDR,
												validation.StringInSlice([]string{`*`}, false),
											),
										},
									},
									"source_ip_groups": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
									"destination_fqdns": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
									"destination_fqdn_tags": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
								},
							},
						},
					},
				},
			},

			"network_rule_collection": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"priority": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(100, 65000),
						},
						"action": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(armnetwork.FirewallPolicyFilterRuleCollectionActionTypeAllow),
								string(armnetwork.FirewallPolicyFilterRuleCollectionActionTypeDeny),
							}, false),
						},
						"rule": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validate.FirewallPolicyRuleName(),
									},
									"protocols": {
										Type:     schema.TypeSet,
										Required: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
											ValidateFunc: validation.StringInSlice([]string{
												string(armnetwork.FirewallPolicyRuleNetworkProtocolAny),
												string(armnetwork.FirewallPolicyRuleNetworkProtocolTCP),
												string(armnetwork.FirewallPolicyRuleNetworkProtocolUDP),
												string(armnetwork.FirewallPolicyRuleNetworkProtocolICMP),
											}, false),
										},
									},
									"source_addresses": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
											ValidateFunc: validation.Any(
												validation.IsIPAddress,
												validation.IsCIDR,
												validation.StringInSlice([]string{`*`}, false),
											),
										},
									},
									"source_ip_groups": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
									"destination_addresses": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
											// Can be IP address, CIDR, "*", or service tag
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
									"destination_ip_groups": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
									"destination_fqdns": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
									"destination_ports": {
										Type:     schema.TypeSet,
										Required: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: azValidate.PortOrPortRangeWithin(1, 65535),
										},
									},
								},
							},
						},
					},
				},
			},

			"nat_rule_collection": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"priority": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(100, 65000),
						},
						"action": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								// Hardcode to using `Dnat` instead of the one defined in Swagger (i.e. network.DNAT) because of: https://github.com/Azure/azure-rest-api-specs/issues/9986
								// Setting `StateFunc: state.IgnoreCase` will cause other issues, as tracked by: https://github.com/hashicorp/terraform-plugin-sdk/issues/485
								// Another solution is to customize the hash function for the containing block, but as there are a couple of properties here, especially
								// has property whose type is another nested block (Set), so the implementation is nontrivial and error-prone.
								"Dnat",
							}, false),
						},
						"rule": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validate.FirewallPolicyRuleName(),
									},
									"protocols": {
										Type:     schema.TypeSet,
										Required: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
											ValidateFunc: validation.StringInSlice([]string{
												string(armnetwork.FirewallPolicyRuleNetworkProtocolTCP),
												string(armnetwork.FirewallPolicyRuleNetworkProtocolUDP),
											}, false),
										},
									},
									"source_addresses": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
											ValidateFunc: validation.Any(
												validation.IsIPAddress,
												validation.IsCIDR,
												validation.StringInSlice([]string{`*`}, false),
											),
										},
									},
									"source_ip_groups": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
									},
									"destination_address": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.Any(
											validation.IsIPAddress,
											validation.IsCIDR,
										),
									},
									"destination_ports": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: azValidate.PortOrPortRangeWithin(1, 64000),
										},
									},
									"translated_address": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.IsIPAddress,
									},
									"translated_port": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IsPortNumber,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceFirewallPolicyRuleCollectionGroupCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Firewall.FirewallPolicyRuleGroupClient2
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	policyId, err := parse.FirewallPolicyID(d.Get("firewall_policy_id").(string))
	if err != nil {
		return err
	}

	if d.IsNewResource() {
		resp, err := client.Get(ctx, policyId.ResourceGroup, policyId.Name, name, nil)
		if err != nil && !utils.Track2ResponseWasNotFound(err) {
			return fmt.Errorf("checking for existing Firewall Policy Rule Collection Group %q (Resource Group %q / Policy %q): %+v", name, policyId.ResourceGroup, policyId.Name, err)
		}
		if err == nil && resp.FirewallPolicyRuleCollectionGroup.ID != nil {
			return tf.ImportAsExistsError("azurerm_firewall_policy_rule_collection_group", *resp.FirewallPolicyRuleCollectionGroup.ID)
		}
	}

	locks.ByName(policyId.Name, azureFirewallPolicyResourceName)
	defer locks.UnlockByName(policyId.Name, azureFirewallPolicyResourceName)

	param := armnetwork.FirewallPolicyRuleCollectionGroup{
		Properties: &armnetwork.FirewallPolicyRuleCollectionGroupProperties{
			Priority: utils.Int32(int32(d.Get("priority").(int))),
		},
	}
	var rulesCollections []armnetwork.FirewallPolicyRuleCollectionClassification
	rulesCollections = append(rulesCollections, expandFirewallPolicyRuleCollectionApplication(d.Get("application_rule_collection").(*schema.Set).List())...)
	rulesCollections = append(rulesCollections, expandFirewallPolicyRuleCollectionNetwork(d.Get("network_rule_collection").(*schema.Set).List())...)
	rulesCollections = append(rulesCollections, expandFirewallPolicyRuleCollectionNat(d.Get("nat_rule_collection").(*schema.Set).List())...)
	param.Properties.RuleCollections = &rulesCollections

	future, err := client.BeginCreateOrUpdate(ctx, policyId.ResourceGroup, policyId.Name, name, param, nil)
	if err != nil {
		return fmt.Errorf("creating Firewall Policy Rule Collection Group %q (Resource Group %q / Policy: %q): %+v", name, policyId.ResourceGroup, policyId.Name, err)
	}
	if _, err := future.PollUntilDone(ctx, time.Minute); err != nil {
		return fmt.Errorf("waiting Firewall Policy Rule Collection Group %q (Resource Group %q / Policy: %q): %+v", name, policyId.ResourceGroup, policyId.Name, err)
	}

	resp, err := client.Get(ctx, policyId.ResourceGroup, policyId.Name, name, nil)
	if err != nil {
		return fmt.Errorf("retrieving Firewall Policy Rule Collection Group %q (Resource Group %q / Policy: %q): %+v", name, policyId.ResourceGroup, policyId.Name, err)
	}
	if resp.FirewallPolicyRuleCollectionGroup.ID == nil || *resp.FirewallPolicyRuleCollectionGroup.ID == "" {
		return fmt.Errorf("empty or nil ID returned for Firewall Policy Rule Collection Group %q (Resource Group %q / Policy: %q) ID", name, policyId.ResourceGroup, policyId.Name)
	}
	id, err := parse.FirewallPolicyRuleCollectionGroupID(*resp.FirewallPolicyRuleCollectionGroup.ID)
	if err != nil {
		return err
	}
	d.SetId(id.ID())

	return resourceFirewallPolicyRuleCollectionGroupRead(d, meta)
}

func resourceFirewallPolicyRuleCollectionGroupRead(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).Firewall.FirewallPolicyRuleGroupClient2
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FirewallPolicyRuleCollectionGroupID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.FirewallPolicyName, id.RuleCollectionGroupName, nil)
	if err != nil {
		if utils.Track2ResponseWasNotFound(err) {
			log.Printf("[DEBUG] Firewall Policy Rule Collection Group %q was not found in Resource Group %q - removing from state!", id.RuleCollectionGroupName, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Firewall Policy Rule Collection Group %q (Resource Group %q / Policy: %q): %+v", id.RuleCollectionGroupName, id.ResourceGroup, id.FirewallPolicyName, err)
	}

	d.Set("name", resp.FirewallPolicyRuleCollectionGroup.Name)

	if prop := resp.FirewallPolicyRuleCollectionGroup.Properties; prop != nil {
		var priority int32
		if prop.Priority != nil {
			priority = *resp.FirewallPolicyRuleCollectionGroup.Properties.Priority
		}
		d.Set("priority", priority)

		d.Set("firewall_policy_id", parse.NewFirewallPolicyID(subscriptionId, id.ResourceGroup, id.FirewallPolicyName).ID())

		applicationRuleCollections, networkRuleCollections, natRuleCollections, err := flattenFirewallPolicyRuleCollection(prop.RuleCollections)
		if err != nil {
			return fmt.Errorf("flattening Firewall Policy Rule Collections: %+v", err)
		}

		if err := d.Set("application_rule_collection", applicationRuleCollections); err != nil {
			return fmt.Errorf("setting `application_rule_collection`: %+v", err)
		}
		if err := d.Set("network_rule_collection", networkRuleCollections); err != nil {
			return fmt.Errorf("setting `network_rule_collection`: %+v", err)
		}
		if err := d.Set("nat_rule_collection", natRuleCollections); err != nil {
			return fmt.Errorf("setting `nat_rule_collection`: %+v", err)
		}
	}

	return nil
}

func resourceFirewallPolicyRuleCollectionGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Firewall.FirewallPolicyRuleGroupClient2
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FirewallPolicyRuleCollectionGroupID(d.Id())
	if err != nil {
		return err
	}

	locks.ByName(id.FirewallPolicyName, azureFirewallPolicyResourceName)
	defer locks.UnlockByName(id.FirewallPolicyName, azureFirewallPolicyResourceName)

	future, err := client.BeginDelete(ctx, id.ResourceGroup, id.FirewallPolicyName, id.RuleCollectionGroupName, nil)
	if err != nil {
		return fmt.Errorf("deleting Firewall Policy Rule Collection Group %q (Resource Group %q / Policy: %q): %+v", id.RuleCollectionGroupName, id.ResourceGroup, id.FirewallPolicyName, err)
	}
	if _, err = future.PollUntilDone(ctx, time.Minute); err != nil {
		return fmt.Errorf("waiting for deleting %q (Resource Group %q / Policy: %q): %+v", id.RuleCollectionGroupName, id.ResourceGroup, id.FirewallPolicyName, err)
	}

	return nil
}

func expandFirewallPolicyRuleCollectionApplication(input []interface{}) []armnetwork.FirewallPolicyRuleCollectionClassification {
	return expandFirewallPolicyFilterRuleCollection(input, expandFirewallPolicyRuleApplication)
}

func expandFirewallPolicyRuleCollectionNetwork(input []interface{}) []armnetwork.FirewallPolicyRuleCollectionClassification {
	return expandFirewallPolicyFilterRuleCollection(input, expandFirewallPolicyRuleNetwork)
}

func expandFirewallPolicyRuleCollectionNat(input []interface{}) []armnetwork.FirewallPolicyRuleCollectionClassification {
	result := make([]armnetwork.FirewallPolicyRuleCollectionClassification, 0)
	for _, e := range input {
		rule := e.(map[string]interface{})
		t := armnetwork.FirewallPolicyRuleCollectionTypeFirewallPolicyNatRuleCollection
		at := armnetwork.FirewallPolicyNatRuleCollectionActionType(rule["action"].(string))
		output := &armnetwork.FirewallPolicyNatRuleCollection{
			FirewallPolicyRuleCollection: armnetwork.FirewallPolicyRuleCollection{
				RuleCollectionType: &t,
				Name:               utils.String(rule["name"].(string)),
				Priority:           utils.Int32(int32(rule["priority"].(int))),
			},
			Action: &armnetwork.FirewallPolicyNatRuleCollectionAction{
				Type: &at,
			},
			Rules: expandFirewallPolicyRuleNat(rule["rule"].(*schema.Set).List()),
		}
		result = append(result, output)
	}
	return result
}

func expandFirewallPolicyFilterRuleCollection(input []interface{}, f func(input []interface{}) *[]armnetwork.FirewallPolicyRuleClassification) []armnetwork.FirewallPolicyRuleCollectionClassification {
	result := make([]armnetwork.FirewallPolicyRuleCollectionClassification, 0)
	for _, e := range input {
		rule := e.(map[string]interface{})
		output := &armnetwork.FirewallPolicyFilterRuleCollection{
			FirewallPolicyRuleCollection: armnetwork.FirewallPolicyRuleCollection{
				Name:               utils.String(rule["name"].(string)),
				Priority:           utils.Int32(int32(rule["priority"].(int))),
				RuleCollectionType: armnetwork.FirewallPolicyRuleCollectionTypeFirewallPolicyFilterRuleCollection.ToPtr(),
			},
			Action: &armnetwork.FirewallPolicyFilterRuleCollectionAction{
				Type: armnetwork.FirewallPolicyFilterRuleCollectionActionType(rule["action"].(string)).ToPtr(),
			},
			Rules: f(rule["rule"].(*schema.Set).List()),
		}
		result = append(result, output)
	}
	return result
}

func expandFirewallPolicyRuleApplication(input []interface{}) *[]armnetwork.FirewallPolicyRuleClassification {
	result := make([]armnetwork.FirewallPolicyRuleClassification, 0)
	for _, e := range input {
		condition := e.(map[string]interface{})
		var protocols []*armnetwork.FirewallPolicyRuleApplicationProtocol
		for _, p := range condition["protocols"].(*schema.Set).List() {
			proto := p.(map[string]interface{})
			protocols = append(protocols, &armnetwork.FirewallPolicyRuleApplicationProtocol{
				ProtocolType: armnetwork.FirewallPolicyRuleApplicationProtocolType(proto["type"].(string)).ToPtr(),
				Port:         utils.Int32(int32(proto["port"].(int))),
			})
		}
		output := &armnetwork.ApplicationRule{
			FirewallPolicyRule: armnetwork.FirewallPolicyRule{
				Name:     utils.String(condition["name"].(string)),
				RuleType: armnetwork.FirewallPolicyRuleTypeApplicationRule.ToPtr(),
			},
			Protocols:       &protocols,
			SourceAddresses: utils.ExpandStringPtrSlice(condition["source_addresses"].(*schema.Set).List()),
			SourceIPGroups:  utils.ExpandStringPtrSlice(condition["source_ip_groups"].(*schema.Set).List()),
			TargetFqdns:     utils.ExpandStringPtrSlice(condition["destination_fqdns"].(*schema.Set).List()),
			FqdnTags:        utils.ExpandStringPtrSlice(condition["destination_fqdn_tags"].(*schema.Set).List()),
		}
		result = append(result, output)
	}
	return &result
}

func expandFirewallPolicyRuleNetwork(input []interface{}) *[]armnetwork.FirewallPolicyRuleClassification {
	result := make([]armnetwork.FirewallPolicyRuleClassification, 0)
	for _, e := range input {
		condition := e.(map[string]interface{})
		var protocols []*armnetwork.FirewallPolicyRuleNetworkProtocol
		for _, p := range condition["protocols"].(*schema.Set).List() {
			protocols = append(protocols, armnetwork.FirewallPolicyRuleNetworkProtocol(p.(string)).ToPtr())
		}
		output := &armnetwork.NetworkRule{
			FirewallPolicyRule: armnetwork.FirewallPolicyRule{
				Name:     utils.String(condition["name"].(string)),
				RuleType: armnetwork.FirewallPolicyRuleTypeNetworkRule.ToPtr(),
			},
			IPProtocols:          &protocols,
			SourceAddresses:      utils.ExpandStringPtrSlice(condition["source_addresses"].(*schema.Set).List()),
			SourceIPGroups:       utils.ExpandStringPtrSlice(condition["source_ip_groups"].(*schema.Set).List()),
			DestinationAddresses: utils.ExpandStringPtrSlice(condition["destination_addresses"].(*schema.Set).List()),
			DestinationIPGroups:  utils.ExpandStringPtrSlice(condition["destination_ip_groups"].(*schema.Set).List()),
			DestinationFqdns:     utils.ExpandStringPtrSlice(condition["destination_fqdns"].(*schema.Set).List()),
			DestinationPorts:     utils.ExpandStringPtrSlice(condition["destination_ports"].(*schema.Set).List()),
		}
		result = append(result, output)
	}
	return &result
}

func expandFirewallPolicyRuleNat(input []interface{}) *[]armnetwork.FirewallPolicyRuleClassification {
	result := make([]armnetwork.FirewallPolicyRuleClassification, 0)
	for _, e := range input {
		condition := e.(map[string]interface{})
		var protocols []*armnetwork.FirewallPolicyRuleNetworkProtocol
		for _, p := range condition["protocols"].(*schema.Set).List() {
			protocols = append(protocols, armnetwork.FirewallPolicyRuleNetworkProtocol(p.(string)).ToPtr())
		}
		destinationAddresses := []*string{utils.String(condition["destination_address"].(string))}
		output := &armnetwork.NatRule{
			FirewallPolicyRule: armnetwork.FirewallPolicyRule{
				Name:     utils.String(condition["name"].(string)),
				RuleType: armnetwork.FirewallPolicyRuleTypeNatRule.ToPtr(),
			},
			IPProtocols:          &protocols,
			SourceAddresses:      utils.ExpandStringPtrSlice(condition["source_addresses"].(*schema.Set).List()),
			SourceIPGroups:       utils.ExpandStringPtrSlice(condition["source_ip_groups"].(*schema.Set).List()),
			DestinationAddresses: &destinationAddresses,
			DestinationPorts:     utils.ExpandStringPtrSlice(condition["destination_ports"].(*schema.Set).List()),
			TranslatedAddress:    utils.String(condition["translated_address"].(string)),
			TranslatedPort:       utils.String(strconv.Itoa(condition["translated_port"].(int))),
		}
		result = append(result, output)
	}
	return &result
}

func flattenFirewallPolicyRuleCollection(input *[]armnetwork.FirewallPolicyRuleCollectionClassification) ([]interface{}, []interface{}, []interface{}, error) {
	var (
		applicationRuleCollection = []interface{}{}
		networkRuleCollection     = []interface{}{}
		natRuleCollection         = []interface{}{}
	)
	if input == nil {
		return applicationRuleCollection, networkRuleCollection, natRuleCollection, nil
	}

	for _, e := range *input {
		var result map[string]interface{}

		switch rule := e.(type) {
		case *armnetwork.FirewallPolicyFilterRuleCollection:
			var name string
			if rule.Name != nil {
				name = *rule.Name
			}
			var priority int32
			if rule.Priority != nil {
				priority = *rule.Priority
			}

			var action string
			if rule.Action != nil && rule.Action.Type != nil {
				action = string(*rule.Action.Type)
			}

			result = map[string]interface{}{
				"name":     name,
				"priority": priority,
				"action":   action,
			}

			if rule.Rules == nil || len(*rule.Rules) == 0 {
				continue
			}

			// Determine the rule type based on the first rule's type
			switch (*rule.Rules)[0].(type) {
			case *armnetwork.ApplicationRule:
				appRules, err := flattenFirewallPolicyRuleApplication(rule.Rules)
				if err != nil {
					return nil, nil, nil, err
				}
				result["rule"] = appRules

				applicationRuleCollection = append(applicationRuleCollection, result)

			case *armnetwork.NetworkRule:
				networkRules, err := flattenFirewallPolicyRuleNetwork(rule.Rules)
				if err != nil {
					return nil, nil, nil, err
				}
				result["rule"] = networkRules

				networkRuleCollection = append(networkRuleCollection, result)

			default:
				return nil, nil, nil, fmt.Errorf("unknown rule type %+v", *(*rule.Rules)[0].GetFirewallPolicyRule().RuleType)
			}
		case *armnetwork.FirewallPolicyNatRuleCollection:
			var name string
			if rule.Name != nil {
				name = *rule.Name
			}
			var priority int32
			if rule.Priority != nil {
				priority = *rule.Priority
			}

			var action string
			if rule.Action != nil && rule.Action.Type != nil {
				action = string(*rule.Action.Type)
			}

			rules, err := flattenFirewallPolicyRuleNat(rule.Rules)
			if err != nil {
				return nil, nil, nil, err
			}
			result = map[string]interface{}{
				"name":     name,
				"priority": priority,
				"action":   action,
				"rule":     rules,
			}

			natRuleCollection = append(natRuleCollection, result)

		default:
			return nil, nil, nil, fmt.Errorf("unknown rule collection type %+v: %v", reflect.TypeOf(rule), spew.Sdump(rule))
		}
	}
	return applicationRuleCollection, networkRuleCollection, natRuleCollection, nil
}

func flattenFirewallPolicyRuleApplication(input *[]armnetwork.FirewallPolicyRuleClassification) ([]interface{}, error) {
	if input == nil {
		return []interface{}{}, nil
	}
	output := make([]interface{}, 0)
	for _, e := range *input {
		rule, ok := e.(*armnetwork.ApplicationRule)
		if !ok {
			return nil, fmt.Errorf("unexpected non-application rule: %+v", e)
		}

		var name string
		if rule.Name != nil {
			name = *rule.Name
		}

		protocols := make([]interface{}, 0)
		if rule.Protocols != nil {
			for _, protocol := range *rule.Protocols {
				var port int
				if protocol.Port != nil {
					port = int(*protocol.Port)
				}
				var t string
				if protocol.ProtocolType != nil {
					t = string(*protocol.ProtocolType)
				}
				protocols = append(protocols, map[string]interface{}{
					"type": t,
					"port": port,
				})
			}
		}

		output = append(output, map[string]interface{}{
			"name":                  name,
			"protocols":             protocols,
			"source_addresses":      utils.FlattenStringPtrSlice(rule.SourceAddresses),
			"source_ip_groups":      utils.FlattenStringPtrSlice(rule.SourceIPGroups),
			"destination_fqdns":     utils.FlattenStringPtrSlice(rule.TargetFqdns),
			"destination_fqdn_tags": utils.FlattenStringPtrSlice(rule.FqdnTags),
		})
	}

	return output, nil
}

func flattenFirewallPolicyRuleNetwork(input *[]armnetwork.FirewallPolicyRuleClassification) ([]interface{}, error) {
	if input == nil {
		return []interface{}{}, nil
	}
	output := make([]interface{}, 0)
	for _, e := range *input {
		rule, ok := e.(*armnetwork.NetworkRule)
		if !ok {
			return nil, fmt.Errorf("unexpected non-network rule: %+v", e)
		}

		var name string
		if rule.Name != nil {
			name = *rule.Name
		}

		protocols := make([]interface{}, 0)
		if rule.IPProtocols != nil {
			for _, protocol := range *rule.IPProtocols {
				if protocol != nil {
					protocols = append(protocols, string(*protocol))
				}
			}
		}

		output = append(output, map[string]interface{}{
			"name":                  name,
			"protocols":             protocols,
			"source_addresses":      utils.FlattenStringPtrSlice(rule.SourceAddresses),
			"source_ip_groups":      utils.FlattenStringPtrSlice(rule.SourceIPGroups),
			"destination_addresses": utils.FlattenStringPtrSlice(rule.DestinationAddresses),
			"destination_ip_groups": utils.FlattenStringPtrSlice(rule.DestinationIPGroups),
			"destination_fqdns":     utils.FlattenStringPtrSlice(rule.DestinationFqdns),
			"destination_ports":     utils.FlattenStringPtrSlice(rule.DestinationPorts),
		})
	}
	return output, nil
}

func flattenFirewallPolicyRuleNat(input *[]armnetwork.FirewallPolicyRuleClassification) ([]interface{}, error) {
	if input == nil {
		return []interface{}{}, nil
	}
	output := make([]interface{}, 0)
	for _, e := range *input {
		rule, ok := e.(*armnetwork.NatRule)
		if !ok {
			return nil, fmt.Errorf("unexpected non-nat rule: %+v", e)
		}

		var name string
		if rule.Name != nil {
			name = *rule.Name
		}

		protocols := make([]interface{}, 0)
		if rule.IPProtocols != nil {
			for _, protocol := range *rule.IPProtocols {
				if protocol != nil {
					protocols = append(protocols, string(*protocol))
				}
			}
		}
		destinationAddr := ""
		if rule.DestinationAddresses != nil && len(*rule.DestinationAddresses) != 0 {
			if (*rule.DestinationAddresses)[0] != nil {
				destinationAddr = *(*rule.DestinationAddresses)[0]
			}
		}

		translatedPort := 0
		if rule.TranslatedPort != nil {
			port, err := strconv.Atoi(*rule.TranslatedPort)
			if err != nil {
				return nil, fmt.Errorf(`The "translatedPort" property is not a valid integer (%s)`, *rule.TranslatedPort)
			}
			translatedPort = port
		}

		output = append(output, map[string]interface{}{
			"name":                name,
			"protocols":           protocols,
			"destination_address": destinationAddr,
			"source_addresses":    utils.FlattenStringPtrSlice(rule.SourceAddresses),
			"source_ip_groups":    utils.FlattenStringPtrSlice(rule.SourceIPGroups),
			"destination_ports":   utils.FlattenStringPtrSlice(rule.DestinationPorts),
			"translated_address":  rule.TranslatedAddress,
			"translated_port":     &translatedPort,
		})
	}
	return output, nil
}
