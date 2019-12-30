package datamigration

import (
	"fmt"
	"reflect"

	"github.com/Azure/azure-sdk-for-go/services/datamigration/mgmt/2018-04-19/datamigration"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func dataSourceArmDataMigrationServiceProject() *schema.Resource {
	buildProjectSqlConnectionInfo := func() *schema.Schema {
		return &schema.Schema{
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"additional_settings": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"authentication": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"data_source": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"encrypt_connection": {
						Type:     schema.TypeBool,
						Computed: true,
					},
					"password": {
						Type:      schema.TypeString,
						Computed:  true,
						Sensitive: true,
					},
					"platform": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"trust_server_certificate": {
						Type:     schema.TypeBool,
						Computed: true,
					},
					"user_name": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		}
	}
	return &schema.Resource{
		Read: dataSourceArmDataMigrationServiceProjectRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateName,
			},

			"service_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateName,
			},

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),

			"location": azure.SchemaLocationForDataSource(),

			"source_platform": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"target_platform": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"source_databases": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"source_connection_info": buildProjectSqlConnectionInfo(),
			"target_connection_info": buildProjectSqlConnectionInfo(),

			"tags": tags.SchemaDataSource(),
		},
	}
}

func dataSourceArmDataMigrationServiceProjectRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataMigration.ProjectsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	serviceName := d.Get("service_name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	resp, err := client.Get(ctx, resourceGroup, serviceName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Error: Data Migration Service Project (Project Name %q / Service Name %q / Group Name %q) was not found", name, serviceName, resourceGroup)
		}
		return fmt.Errorf("Error reading Data Migration Service Project (Project Name %q / Service Name %q / Group Name %q): %+v", name, serviceName, resourceGroup, err)
	}

	d.SetId(*resp.ID)

	d.Set("resource_group_name", resourceGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}
	if projectProperties := resp.ProjectProperties; projectProperties != nil {
		if err := d.Set("source_databases", flattenProjectDatabaseInfo(projectProperties.DatabasesInfo)); err != nil {
			return fmt.Errorf("Error setting `source_databases`: %+v", err)
		}
		d.Set("source_platform", string(projectProperties.SourcePlatform))
		d.Set("target_platform", string(projectProperties.TargetPlatform))

		var sourceConnectionInfo []interface{}
		switch projectProperties.SourceConnectionInfo.(type) {
		case datamigration.SQLConnectionInfo:
			v := projectProperties.SourceConnectionInfo.(datamigration.SQLConnectionInfo)
			sourceConnectionInfo = flattenProjectSqlConnectionInfo(&v)
		default:
			return fmt.Errorf("Unknown source connection info: %s", reflect.TypeOf(projectProperties.SourceConnectionInfo))
		}
		if err := d.Set("source_conection_info", sourceConnectionInfo); err != nil {
			return fmt.Errorf("Error setting `source_connection_info`: %+v", err)
		}

		var targetConnectionInfo []interface{}
		switch projectProperties.TargetConnectionInfo.(type) {
		case datamigration.SQLConnectionInfo:
			v := projectProperties.TargetConnectionInfo.(datamigration.SQLConnectionInfo)
			targetConnectionInfo = flattenProjectSqlConnectionInfo(&v)
		default:
			return fmt.Errorf("Unknown target connection info: %s", reflect.TypeOf(projectProperties.TargetConnectionInfo))
		}

		if err := d.Set("target_connection_info", targetConnectionInfo); err != nil {
			return fmt.Errorf("Error setting `target_connection_info`: %+v", err)
		}
	}
	d.Set("id", resp.ID)
	d.Set("type", resp.Type)

	return nil
}
