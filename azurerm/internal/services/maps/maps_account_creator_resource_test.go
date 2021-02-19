package maps

type MapsAccountCreatorResource struct{}

func TestAccAzureRMMapsAccountCreatorc_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_maps_account_creatorc", "test")
	r := MapsAccountCreatorcResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAzureRMMapsAccountCreatorc_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_maps_account_creatorc", "test")
	r := MapsAccountCreatorcResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAzureRMMapsAccountCreatorc_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_maps_account_creatorc", "test")
	r := MapsAccountCreatorcResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAzureRMMapsAccountCreatorc_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_maps_account_creatorc", "test")
	r := MapsAccountCreatorcResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func (r MapsAccountCreatorcResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
	client := clients..Client

	id, err := parse.MapsAccountCreatorcID(state.ID)
	if err != nil {
		return nil, err
	}

	if resp, err := client.Get(ctx, id.ResourceGroup, id.MapsAccountCreatorcName); err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("retrieving Maps Account Creatorc %q: %+v", id, err)
	}

	return utils.Bool(true), nil
}

func (r MapsAccountCreatorcResource) basic(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_maps_account_creatorc" "test" {
}
`, template)
}

func (r MapsAccountCreatorcResource) complete(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_maps_account_creatorc" "test" {
}
`, template)
}

func (r MapsAccountCreatorcResource) requiresImport(data acceptance.TestData) string {
	template := r.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_maps_account_creatorc" "import" {
}
`, template)
}

func (r MapsAccountCreatorcResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
`)
}
