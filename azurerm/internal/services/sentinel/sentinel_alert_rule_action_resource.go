package sentinel

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/securityinsight/mgmt/2019-01-01-preview/securityinsight"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	logicParse "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/logic/parse"
	logicValidate "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/logic/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/sentinel/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/sentinel/validate"
	azSchema "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

const OperationInsightsRPName = "Microsoft.OperationalInsights"

func resourceSentinelAlertRuleAction() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmSentinelAlertRuleActionCreate,
		Read:   resourceArmSentinelAlertRuleActionRead,
		Delete: resourceArmSentinelAlertRuleActionDelete,

		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := parse.ActionID(id)
			return err
		}),

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"sentinel_alert_rule_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.AlertRuleID,
			},

			"logic_app_trigger_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: logicValidate.TriggerID,
			},
		},
	}
}

func resourceArmSentinelAlertRuleActionCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Sentinel.AlertRulesClient
	logicTriggerClient := meta.(*clients.Client).Logic.WorkflowTriggersClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)

	ruleID, err := parse.AlertRuleID(d.Get("sentinel_alert_rule_id").(string))
	if err != nil {
		return err
	}
	triggerId, err := logicParse.TriggerID(d.Get("logic_app_trigger_id").(string))
	if err != nil {
		return err
	}

	// Ensure no existed resources
	resp, err := client.GetAction(ctx, ruleID.ResourceGroup, OperationInsightsRPName, ruleID.WorkspaceName, ruleID.Name, name)
	if err != nil {
		if !utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("checking for existing Sentinel Alert Rule Action %q (%s): %+v", name, ruleID, err)
		}
	}

	if resp.ID != nil && *resp.ID != "" {
		id, err := parse.ActionID(*resp.ID)
		if err != nil {
			return err
		}
		return tf.ImportAsExistsError("azurerm_sentinel_alert_rule_action", id.ID())
	}

	// List callback URL for sentinel alert specific trigger from the workspace containing specified alert rule.
	tresp, err := logicTriggerClient.ListCallbackURL(ctx, ruleID.ResourceGroup, triggerId.WorkflowName, triggerId.Name)
	if err != nil {
		return fmt.Errorf("listing callback URL for Logic App Trigger %q: %v", triggerId, err)
	}

	lappWorkflowId := logicParse.NewWorkflowID(triggerId.SubscriptionId, triggerId.ResourceGroup, triggerId.WorkflowName)

	param := securityinsight.ActionRequest{
		ActionRequestProperties: &securityinsight.ActionRequestProperties{
			TriggerURI:         tresp.Value,
			LogicAppResourceID: utils.String(lappWorkflowId.ID()),
		},
	}

	if _, err := client.CreateOrUpdateAction(ctx, ruleID.ResourceGroup, OperationInsightsRPName, ruleID.WorkspaceName, ruleID.Name, name, param); err != nil {
		return fmt.Errorf("creating Sentinel Alert Rule Action %q (%s): %+v", name, ruleID, err)
	}

	resp, err = client.GetAction(ctx, ruleID.ResourceGroup, OperationInsightsRPName, ruleID.WorkspaceName, ruleID.Name, name)
	if err != nil {
		return fmt.Errorf("retrieving Sentinel Alert Rule Action %q (%s): %+v", name, ruleID, err)
	}
	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("empty or nil ID returned for Sentinel Alert Rule Action %q (%s)", name, ruleID)
	}

	id, err := parse.ActionID(*resp.ID)
	if err != nil {
		return err
	}
	d.SetId(id.ID())

	return resourceArmSentinelAlertRuleActionRead(d, meta)
}

func resourceArmSentinelAlertRuleActionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Sentinel.AlertRulesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ActionID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.GetAction(ctx, id.ResourceGroup, OperationInsightsRPName, id.WorkspaceName, id.AlertRuleName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Sentinel Alert Rule Action %q was not found - removing from state!", id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Sentinel Alert Rule Action %q: %+v", id, err)
	}

	d.Set("name", id.Name)

	ruleID := parse.NewAlertRuleID(id.SubscriptionId, id.ResourceGroup, id.WorkspaceName, id.AlertRuleName)
	d.Set("sentinel_alert_rule_id", ruleID.ID())

	if prop := resp.ActionResponseProperties; prop != nil {
		// TODO: Set trigger id once https://github.com/Azure/azure-rest-api-specs/issues/9424 is addressed.
	}

	return nil
}

func resourceArmSentinelAlertRuleActionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Sentinel.AlertRulesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ActionID(d.Id())
	if err != nil {
		return err
	}

	if _, err := client.DeleteAction(ctx, id.ResourceGroup, OperationInsightsRPName, id.WorkspaceName, id.AlertRuleName, id.Name); err != nil {
		return fmt.Errorf("deleting Sentinel Alert Rule Action %q: %+v", id, err)
	}

	return nil
}
