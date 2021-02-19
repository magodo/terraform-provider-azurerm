package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
)

type CreatorId struct {
	SubscriptionId string
	ResourceGroup  string
	AccountName    string
	Name           string
}

func NewCreatorID(subscriptionId, resourceGroup, accountName, name string) CreatorId {
	return CreatorId{
		SubscriptionId: subscriptionId,
		ResourceGroup:  resourceGroup,
		AccountName:    accountName,
		Name:           name,
	}
}

func (id CreatorId) String() string {
	segments := []string{
		fmt.Sprintf("Name %q", id.Name),
		fmt.Sprintf("Account Name %q", id.AccountName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Creator", segmentsStr)
}

func (id CreatorId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Maps/accounts/%s/creators/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.AccountName, id.Name)
}

// CreatorID parses a Creator ID into an CreatorId struct
func CreatorID(input string) (*CreatorId, error) {
	id, err := azure.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := CreatorId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.AccountName, err = id.PopSegment("accounts"); err != nil {
		return nil, err
	}
	if resourceId.Name, err = id.PopSegment("creators"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
