package migration

import (
	"context"

	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

var _ pluginsdk.StateUpgrade = VirtualNetworkV0ToV1{}

type VirtualNetworkV0ToV1 struct{}

func (VirtualNetworkV0ToV1) Schema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},

		"resource_group_name": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},

		"location": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},

		"address_space": {
			Type:     pluginsdk.TypeSet,
			Required: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},

		"bgp_community": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},

		"ddos_protection_plan": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"id": {
						Type:     pluginsdk.TypeString,
						Required: true,
					},

					"enable": {
						Type:     pluginsdk.TypeBool,
						Required: true,
					},
				},
			},
		},

		"encryption": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"enforcement": {
						Type:     pluginsdk.TypeString,
						Required: true,
					},
				},
			},
		},

		"dns_servers": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Computed: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},

		"edge_zone": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			ForceNew: true,
		},

		"flow_timeout_in_minutes": {
			Type:     pluginsdk.TypeInt,
			Optional: true,
		},

		"guid": {
			Type:     pluginsdk.TypeString,
			Computed: true,
		},

		"subnet": {
			Type:       pluginsdk.TypeSet,
			Optional:   true,
			Computed:   true,
			ConfigMode: pluginsdk.SchemaConfigModeAttr,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"name": {
						Type:     pluginsdk.TypeString,
						Required: true,
					},

					"address_prefixes": {
						Type:     pluginsdk.TypeList,
						Required: true,
						MinItems: 1,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
					},

					"default_outbound_access_enabled": {
						Type:     pluginsdk.TypeBool,
						Default:  true,
						Optional: true,
					},

					"delegation": {
						Type:       pluginsdk.TypeList,
						Optional:   true,
						ConfigMode: pluginsdk.SchemaConfigModeAttr,
						Elem: &pluginsdk.Resource{
							Schema: map[string]*pluginsdk.Schema{
								"name": {
									Type:     pluginsdk.TypeString,
									Required: true,
								},
								"service_delegation": {
									Type:       pluginsdk.TypeList,
									Required:   true,
									ConfigMode: pluginsdk.SchemaConfigModeAttr,
									Elem: &pluginsdk.Resource{
										Schema: map[string]*pluginsdk.Schema{
											"name": {
												Type:     pluginsdk.TypeString,
												Required: true,
											},

											"actions": {
												Type:     pluginsdk.TypeSet,
												Optional: true,
												Elem: &pluginsdk.Schema{
													Type: pluginsdk.TypeString,
												},
											},
										},
									},
								},
							},
						},
					},

					"private_endpoint_network_policies": {
						Type:     pluginsdk.TypeString,
						Optional: true,
					},

					"private_link_service_network_policies_enabled": {
						Type:     pluginsdk.TypeBool,
						Optional: true,
						Default:  true,
					},

					"route_table_id": {
						Type:     pluginsdk.TypeString,
						Optional: true,
					},

					"security_group": {
						Type:     pluginsdk.TypeString,
						Optional: true,
					},

					"service_endpoints": {
						Type:     pluginsdk.TypeSet,
						Optional: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
						Set: pluginsdk.HashString,
					},

					"service_endpoint_policy_ids": {
						Type:     pluginsdk.TypeSet,
						Optional: true,
						Elem: &pluginsdk.Schema{
							Type: pluginsdk.TypeString,
						},
					},

					"id": {
						Type:     pluginsdk.TypeString,
						Computed: true,
					},
				},
			},
		},

		"tags": {
			Type:     pluginsdk.TypeMap,
			Optional: true,
			Elem: &pluginsdk.Schema{
				Type: pluginsdk.TypeString,
			},
		},
	}
}

func (VirtualNetworkV0ToV1) UpgradeFunc() pluginsdk.StateUpgraderFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		rawState["locations"] = []interface{}{rawState["location"]}
		delete(rawState, "location")

		rawState["uuid"] = rawState["guid"]
		delete(rawState, "guid")

		return rawState, nil
	}
}
