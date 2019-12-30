package datamigration

import (
	"fmt"
	"log"
	"reflect"

	"github.com/Azure/azure-sdk-for-go/services/datamigration/mgmt/2018-04-19/datamigration"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmDataMigrationServiceProject() *schema.Resource {

	buildProjectSqlConnectionInfo := func() *schema.Schema {
		return &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"additional_settings": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validate.NoEmptyStrings,
					},
					"authentication": {
						Type:     schema.TypeString,
						Optional: true,
						ValidateFunc: validation.StringInSlice(
							[]string{
								string(datamigration.ActiveDirectoryIntegrated),
								string(datamigration.ActiveDirectoryPassword),
								string(datamigration.None),
								string(datamigration.SQLAuthentication),
								string(datamigration.WindowsAuthentication),
							},
							false,
						),
					},
					"data_source": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validateSqlConnectionInfoSourceName,
					},
					"encrypt_connection": {
						Type:     schema.TypeBool,
						Optional: true,
					},
					"password": {
						Type:         schema.TypeString,
						Optional:     true,
						Sensitive:    true,
						ValidateFunc: validate.NoEmptyStrings,
					},
					"platform": {
						Type:     schema.TypeString,
						Optional: true,
						ValidateFunc: validation.StringInSlice(
							[]string{string(datamigration.SQLOnPrem)},
							false,
						),
					},
					"trust_server_certificate": {
						Type:     schema.TypeBool,
						Optional: true,
					},
					"user_name": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validate.NoEmptyStrings,
					},
				},
			},
		}
	}

	return &schema.Resource{
		Create: resourceArmDataMigrationServiceProjectCreateUpdate,
		Read:   resourceArmDataMigrationServiceProjectRead,
		Update: resourceArmDataMigrationServiceProjectCreateUpdate,
		Delete: resourceArmDataMigrationServiceProjectDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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

			"resource_group_name": azure.SchemaResourceGroupNameDiffSuppress(),

			"location": azure.SchemaLocation(),

			"source_platform": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					// Now that go sdk only export SQL as source platform type, we only allow it here.
					string(datamigration.ProjectSourcePlatformSQL),
				}, false),
			},

			"target_platform": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					// Now that go sdk only export SQL as source platform type, we only allow it here.
					string(datamigration.ProjectTargetPlatformSQLDB),
				}, false),
			},

			"source_databases": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"sql_source_connection_info": buildProjectSqlConnectionInfo(),
			"sql_target_connection_info": buildProjectSqlConnectionInfo(),

			"tags": tags.Schema(),
		},
	}
}

func resourceArmDataMigrationServiceProjectCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataMigration.ProjectsClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	serviceName := d.Get("service_name").(string)

	if features.ShouldResourcesBeImported() && d.IsNewResource() {
		existing, err := client.Get(ctx, resourceGroup, serviceName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for present of existing Data Migration Service Project (Project Name: %q / Service Name %q / Group Name %q): %+v", name, serviceName, resourceGroup, err)
			}
		}
		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_data_migration_service_project", *existing.ID)
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	sourcePlatform := d.Get("source_platform").(string)
	targetPlatform := d.Get("target_platform").(string)
	t := d.Get("tags").(map[string]interface{})

	parameters := datamigration.Project{
		Location: utils.String(location),
		ProjectProperties: &datamigration.ProjectProperties{
			SourcePlatform: datamigration.ProjectSourcePlatform(sourcePlatform),
			TargetPlatform: datamigration.ProjectTargetPlatform(targetPlatform),
		},
		Tags: tags.Expand(t),
	}
	if sourceDatabases, ok := d.GetOk("source_databases"); ok {
		parameters.ProjectProperties.DatabasesInfo = expandProjectDatabaseInfo(sourceDatabases.(*schema.Set).List())
	}

	if sourceConnectionInfo, ok := d.GetOk("source_connection_info"); ok {
		switch datamigration.ProjectSourcePlatform(sourcePlatform) {
		case datamigration.ProjectSourcePlatformSQL:
			parameters.SourceConnectionInfo = expandProjectSqlConnectionInfo(sourceConnectionInfo.([]interface{}))
		default:
			fmt.Errorf("Unknown source platform: %s", sourcePlatform)
		}
	}

	if targetConnectionInfo, ok := d.GetOk("target_connection_info"); ok {
		switch datamigration.ProjectTargetPlatform(targetPlatform) {
		case datamigration.ProjectTargetPlatformSQLDB:
			parameters.TargetConnectionInfo = expandProjectSqlConnectionInfo(targetConnectionInfo.([]interface{}))
		default:
			fmt.Errorf("Unknown target platform: %s", targetPlatform)
		}
	}

	if _, err := client.CreateOrUpdate(ctx, parameters, resourceGroup, serviceName, name); err != nil {
		return fmt.Errorf("Error creating Data Migration Service Project (Project Name %q / Service Name %q / Group Name %q): %+v", name, serviceName, resourceGroup, err)
	}

	resp, err := client.Get(ctx, resourceGroup, serviceName, name)
	if err != nil {
		return fmt.Errorf("Error retrieving Data Migration Service Project (Project Name %q / Service Name %q / Group Name %q): %+v", name, serviceName, resourceGroup, err)
	}
	if resp.ID == nil {
		return fmt.Errorf("Cannot read Data Migration Service Project (Project Name %q / Service Name %q / Group Name %q) ID", name, serviceName, resourceGroup)
	}
	d.SetId(*resp.ID)

	return resourceArmDataMigrationServiceProjectRead(d, meta)
}

func resourceArmDataMigrationServiceProjectRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataMigration.ProjectsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	name := id.Path["projects"]
	resourceGroup := id.ResourceGroup
	serviceName := id.Path["services"]

	resp, err := client.Get(ctx, resourceGroup, serviceName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Data Migration Service Project %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading Data Migration Service Project (Project Name %q / Service Name %q / Group Name %q): %+v", name, serviceName, resourceGroup, err)
	}

	d.Set("name", resp.Name)
	d.Set("service_name", serviceName)
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

		if projectProperties.SourceConnectionInfo != nil {
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
		}

		if projectProperties.TargetConnectionInfo != nil {
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
	}
	d.Set("id", resp.ID)

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceArmDataMigrationServiceProjectDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataMigration.ProjectsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	serviceName := id.Path["services"]
	name := id.Path["projects"]

	deleteRunningTasks := true
	if _, err := client.Delete(ctx, resourceGroup, serviceName, name, &deleteRunningTasks); err != nil {
		return fmt.Errorf("Error deleting Data Migration Service Project (Project Name %q / Service Name %q / Group Name %q): %+v", name, serviceName, resourceGroup, err)
	}

	return nil
}

func expandProjectDatabaseInfo(input []interface{}) *[]datamigration.DatabaseInfo {
	results := make([]datamigration.DatabaseInfo, 0)
	for _, item := range input {
		dbName := item.(string)

		result := datamigration.DatabaseInfo{
			SourceDatabaseName: utils.String(dbName),
		}

		results = append(results, result)
	}
	return &results
}

func expandProjectSqlConnectionInfo(input []interface{}) *datamigration.SQLConnectionInfo {
	if len(input) == 0 {
		return nil
	}
	v := input[0].(map[string]interface{})
	result := datamigration.SQLConnectionInfo{}

	if additionalSettings, ok := v["additional_settings"]; ok {
		result.AdditionalSettings = utils.String(additionalSettings.(string))
	}

	if authentication, ok := v["authentication"]; ok {
		result.Authentication = datamigration.AuthenticationType(authentication.(string))
	}

	if dataSource, ok := v["data_source"]; ok {
		result.DataSource = utils.String(dataSource.(string))
	}

	if encryptConnection, ok := v["encrypt_connection"]; ok {
		result.EncryptConnection = utils.Bool(encryptConnection.(bool))
	}

	if password, ok := v["password"]; ok {
		result.Password = utils.String(password.(string))
	}

	if platform, ok := v["platform"]; ok {
		result.Platform = datamigration.SQLSourcePlatform(platform.(string))
	}

	if trustServerCertificate, ok := v["trust_server_certificate"]; ok {
		result.TrustServerCertificate = utils.Bool(trustServerCertificate.(bool))
	}

	if userName, ok := v["user_name"]; ok {
		result.UserName = utils.String(userName.(string))
	}

	// Explicitly specify type of connection info. User has already choose the type via filling in the corresponding xxx_connection_info property
	result.Type = "SqlConnectionInfo"

	return &result
}

func flattenProjectDatabaseInfo(input *[]datamigration.DatabaseInfo) *schema.Set {
	results := make([]interface{}, 0)
	if input == nil {
		return schema.NewSet(schema.HashString, results)
	}

	for _, item := range *input {
		if dbName := item.SourceDatabaseName; dbName != nil {
			results = append(results, *dbName)
		}
	}

	return schema.NewSet(schema.HashString, results)
}

func flattenProjectSqlConnectionInfo(input *datamigration.SQLConnectionInfo) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	result := make(map[string]interface{})

	if additionalSettings := input.AdditionalSettings; additionalSettings != nil {
		result["additional_settings"] = *additionalSettings
	}

	if authentication := input.Authentication; authentication != "" {
		result["authentication"] = string(authentication)
	}

	if dataSource := input.DataSource; dataSource != nil {
		result["data_source"] = *dataSource
	}

	if encryptConnection := input.EncryptConnection; encryptConnection != nil {
		result["encrypt_connection"] = *encryptConnection
	}

	if password := input.Password; password != nil {
		result["password"] = *password
	}

	if platform := input.Platform; platform != "" {
		result["platform"] = string(platform)
	}

	if trustServerCertificate := input.TrustServerCertificate; trustServerCertificate != nil {
		result["trust_server_certificate"] = *trustServerCertificate
	}

	if userName := input.UserName; userName != nil {
		result["user_name"] = *userName
	}

	return []interface{}{result}
}
