package datamigration

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/datamigration/mgmt/2018-04-19/datamigration"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

const sqlServerDefaultPort = 1433

func resourceArmDataMigrationServiceTaskSqlServer() *schema.Resource {
	buildProjectSqlConnectionInfo := func() *schema.Schema {
		return &schema.Schema{
			Type:     schema.TypeList,
			ForceNew: true,
			Required: true,
			MaxItems: 1,
			MinItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"server_address": {
						Type:     schema.TypeString,
						Required: true,
						ForceNew: true,
					},
					"server_port": {
						Type:     schema.TypeInt,
						Optional: true,
						ForceNew: true,
						Default:  sqlServerDefaultPort,
					},
					"user_name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"password": {
						Type:      schema.TypeString,
						Optional:  true,
						Sensitive: true,
					},
					"encrypt_connection": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  true,
					},
					"trust_server_certificate": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  true,
					},
				},
			},
		}
	}

	return &schema.Resource{
		Create: resourceArmDataMigrationServiceTaskSqlServerCreate,
		Read:   resourceArmDataMigrationServiceTaskSqlServerRead,
		Delete: resourceArmDataMigrationServiceTaskSqlServerDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateTaskName,
			},

			"project_name": {
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

			"resource_group_name": azure.SchemaResourceGroupName(),

			"databases": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_database_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"target_database_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"source_tables": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"target_tables": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"source_connection_info": buildProjectSqlConnectionInfo(),
			"target_connection_info": buildProjectSqlConnectionInfo(),
		},
	}
}

func resourceArmDataMigrationServiceTaskSqlServerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataMigration.TasksClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceGroup := d.Get("resource_group_name").(string)
	serviceName := d.Get("service_name").(string)
	projectName := d.Get("project_name").(string)
	name := d.Get("name").(string)

	if features.ShouldResourcesBeImported() && d.IsNewResource() {
		existing, err := client.Get(ctx, resourceGroup, serviceName, projectName, name, "")
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for present of existing Data Migration Service Task (Task Name %q / Project Name %q / Service Name %q / Group Name %q): %+v", name, projectName, serviceName, resourceGroup, err)
			}
		}
		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_data_migration_service_task", *existing.ID)
		}
	}

	databases, err := expandDataMigrationServiceTaskDatabases(d.Get("databases").(*schema.Set))
	if err != nil {
		return fmt.Errorf("Error parsing task databases: %v", err)
	}
	parameters := datamigration.ProjectTask{
		Properties: datamigration.MigrateSQLServerSQLDbSyncTaskProperties{
			TaskType: datamigration.TaskTypeMigrateSQLServerAzureSQLDbSync,
			Input: &datamigration.MigrateSQLServerSQLDbSyncTaskInput{
				SelectedDatabases:    databases,
				SourceConnectionInfo: expandDataMigrationServiceTaskConnectionInfo(d.Get("source_connection_info").([]interface{})),
				TargetConnectionInfo: expandDataMigrationServiceTaskConnectionInfo(d.Get("target_connection_info").([]interface{})),
			},
		},
	}

	// TODO: validate stuffs

	_, err = client.CreateOrUpdate(ctx, parameters, resourceGroup, serviceName, projectName, name)
	if err != nil {
		return fmt.Errorf("Error creating Data Migration Service Task (Task Name %q / Project Name %q / Service Name %q / Group Name %q): %+v", name, projectName, serviceName, resourceGroup, err)
	}

	resp, err := client.Get(ctx, resourceGroup, serviceName, projectName, name, "")
	if err != nil {
		return fmt.Errorf("Error retrieving Data Migration Service Task (Task Name %q / Project Name %q / Service Name %q / Group Name %q): %+v", name, projectName, serviceName, resourceGroup, err)
	}
	if resp.ID == nil {
		return fmt.Errorf("Cannot read Data Migration Service Task (Task Name %q / Project Name %q / Service Name %q / Group Name %q): %+v", name, projectName, serviceName, resourceGroup, err)
	}
	d.SetId(*resp.ID)

	return resourceArmDataMigrationServiceTaskSqlServerRead(d, meta)
}

func resourceArmDataMigrationServiceTaskSqlServerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataMigration.TasksClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup := id.ResourceGroup
	serviceName := id.Path["services"]
	projectName := id.Path["projects"]
	name := id.Path["tasks"]

	resp, err := client.Get(ctx, resourceGroup, serviceName, projectName, name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Data Migration Service Task %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading Data Migration Service Task (Task Name %q / Project Name %q / Service Name %q / Group Name %q): %+v", name, projectName, serviceName, resourceGroup, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resourceGroup)
	d.Set("service_name", serviceName)
	d.Set("project_name", projectName)

	prop, ok := resp.Properties.AsMigrateSQLServerSQLDbSyncTaskProperties()
	if !ok {
		return fmt.Errorf("Property in response is not of expected type: %s", reflect.TypeOf(resp.Properties))
	}
	if err := d.Set("databases", flattenDataMigrationServiceTaskDatabases(prop.Input.SelectedDatabases)); err != nil {
		return fmt.Errorf("Error setting `databases`: %v", err)
	}

	sourceConnectionInfo, err := flattenDataMigrationServiceTaskConnectionInfo(d.Get("source_connection_info.0.password").(string), prop.Input.SourceConnectionInfo)
	if err != nil {
		return fmt.Errorf("Error flattening `source_connection_info`: %v", err)
	}
	if err := d.Set("source_connection_info", sourceConnectionInfo); err != nil {
		return fmt.Errorf("Error setting `source_connection_info`: %v", err)
	}

	targetConnectionInfo, err := flattenDataMigrationServiceTaskConnectionInfo(d.Get("target_connection_info.0.password").(string), prop.Input.TargetConnectionInfo)
	if err != nil {
		return fmt.Errorf("Error flattening `target_connection_info`: %v", err)
	}
	if err := d.Set("target_connection_info", targetConnectionInfo); err != nil {
		return fmt.Errorf("Error setting `target_connection_info`: %v", err)
	}

	return nil
}

func resourceArmDataMigrationServiceTaskSqlServerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataMigration.TasksClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup := id.ResourceGroup
	serviceName := id.Path["services"]
	projectName := id.Path["projects"]
	name := id.Path["tasks"]

	if _, err := client.Delete(ctx, resourceGroup, serviceName, projectName, name, utils.Bool(true)); err != nil {
		return fmt.Errorf("Error deleting Data Migration Service Task (Task Name %q / Project Name %q / Service Name %q / Group Name %q): %+v", name, projectName, serviceName, resourceGroup, err)
	}
	return nil
}

func expandDataMigrationServiceTaskDatabases(input *schema.Set) (*[]datamigration.MigrateSQLServerSQLDbSyncDatabaseInput, error) {
	output := make([]datamigration.MigrateSQLServerSQLDbSyncDatabaseInput, 0)
	for _, b := range input.List() {
		databaseInfo := b.(map[string]interface{})
		var d datamigration.MigrateSQLServerSQLDbSyncDatabaseInput
		d.Name = utils.String(databaseInfo["source_database_name"].(string))
		d.TargetDatabaseName = utils.String(databaseInfo["target_database_name"].(string))
		d.TableMap = map[string]*string{}
		sourceTables := *utils.ExpandStringSlice(databaseInfo["source_tables"].([]interface{}))
		targetTables := *utils.ExpandStringSlice(databaseInfo["target_tables"].([]interface{}))
		if len(sourceTables) != len(targetTables) {
			return nil, fmt.Errorf("Amount of source tables(%d) not equal to target tables(%d)", len(sourceTables), len(targetTables))
		}
		for idx := range sourceTables {
			d.TableMap[sourceTables[idx]] = &targetTables[idx]
		}
		output = append(output, d)
	}
	return &output, nil
}

func flattenDataMigrationServiceTaskDatabases(input *[]datamigration.MigrateSQLServerSQLDbSyncDatabaseInput) *schema.Set {
	outputList := make([]interface{}, 0)
	for _, databaseInfo := range *input {
		d := make(map[string]interface{})
		d["source_database_name"] = *databaseInfo.Name
		d["target_database_name"] = *databaseInfo.TargetDatabaseName
		d["source_tables"] = make([]string, 0)
		d["target_tables"] = make([]string, 0)
		for st, tt := range databaseInfo.TableMap {
			d["source_tables"] = append(d["source_tables"].([]string), st)
			d["target_tables"] = append(d["target_tables"].([]string), *tt)
		}
		outputList = append(outputList, d)
	}
	return schema.NewSet(resourceAzureRMDataMigrationServiceTaskDatabasesHash, outputList)
}

func expandDataMigrationServiceTaskConnectionInfo(input []interface{}) *datamigration.SQLConnectionInfo {
	if len(input) == 0 {
		return nil
	}
	v := input[0].(map[string]interface{})
	return &datamigration.SQLConnectionInfo{
		DataSource:             utils.String(fmt.Sprintf("%s,%d", v["server_address"].(string), v["server_port"].(int))),
		EncryptConnection:      utils.Bool(v["encrypt_connection"].(bool)),
		TrustServerCertificate: utils.Bool(v["trust_server_certificate"].(bool)),
		UserName:               utils.String(v["user_name"].(string)),
		Password:               utils.String(v["password"].(string)),
		Authentication:         datamigration.SQLAuthentication,
		Type:                   datamigration.TypeSQLConnectionInfo,
	}
}

func flattenDataMigrationServiceTaskConnectionInfo(passwordInState string, input *datamigration.SQLConnectionInfo) ([]interface{}, error) {
	if input == nil {
		return []interface{}{}, nil
	}

	addrComponents := strings.Split(*input.DataSource, ",")
	var (
		err        error
		serverAddr string
		serverPort int
	)
	if len(addrComponents) == 1 {
		serverAddr = addrComponents[0]
		serverPort = sqlServerDefaultPort
	} else {
		serverAddr = addrComponents[0]
		serverPort, err = strconv.Atoi(addrComponents[1])
		if err != nil {
			return nil, fmt.Errorf("Error parsing server port string literal(%s) into int: %v", addrComponents[1], err)
		}
	}

	return []interface{}{
		map[string]interface{}{
			"server_address":           serverAddr,
			"server_port":              serverPort,
			"user_name":                *input.UserName,
			"encrypt_connection":       *input.EncryptConnection,
			"trust_server_certificate": *input.TrustServerCertificate,
			// Not flatten "password" from service response since service will return a hash of the password with a random seed,
			// in which case we are not able to detect configuration differences.
			// In stead, we will keep it as what it is in current state
			"password": passwordInState,
		},
	}, nil
}

func resourceAzureRMDataMigrationServiceTaskDatabasesHash(v interface{}) int {
	var buf bytes.Buffer

	if m, ok := v.(map[string]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%s-", m["source_database_name"].(string)))
		buf.WriteString(fmt.Sprintf("%s-", m["target_database_name"].(string)))
		buf.WriteString(fmt.Sprintf("%s-", strings.Join(m["source_tables"].([]string), ",")))
		buf.WriteString(fmt.Sprintf("%s-", strings.Join(m["target_tables"].([]string), ",")))
	}

	return hashcode.String(buf.String())
}
