package datamigration

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func dataSourceArmDataMigrationServiceProject() *schema.Resource {
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
		d.Set("source_platform", string(projectProperties.SourcePlatform))
		d.Set("target_platform", string(projectProperties.TargetPlatform))
	}
	d.Set("id", resp.ID)
	d.Set("type", resp.Type)

	return nil
}
