package parse

import (
	"testing"
)

func TestLogicAppTriggerID(t *testing.T) {
	testData := []struct {
		Name   string
		Input  string
		Error  bool
		Expect *LogicAppTriggerId
	}{
		{
			Name:  "Empty",
			Input: "",
			Error: true,
		},
		{
			Name:  "No Resource Groups Segment",
			Input: "/subscriptions/00000000-0000-0000-0000-000000000000",
			Error: true,
		},
		{
			Name:  "No Resource Groups Value",
			Input: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/",
			Error: true,
		},
		{
			Name:  "No Workflow Name",
			Input: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resGroup1/workflows",
			Error: true,
		},
		{
			Name:  "No Trigger Name",
			Input: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resGroup1/workflows/workflow1/triggers",
			Error: true,
		},
		{
			Name:  "Correct case",
			Input: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resGroup1/workflows/workflow1/triggers/trigger1",
			Expect: &LogicAppTriggerId{
				LogicAppWorkflowId: LogicAppWorkflowId{
					Subscription:  "00000000-0000-0000-0000-000000000000",
					ResourceGroup: "resGroup1",
					Name:          "workflow1",
				},
				Name: "trigger1",
			},
		},
	}

	for _, v := range testData {
		t.Logf("[DEBUG] Testing %q", v.Name)

		actual, err := LogicAppTriggerID(v.Input)
		if err != nil {
			if v.Error {
				continue
			}

			t.Fatalf("Expected a value but got an error: %s", err)
		}

		if actual.Subscription != v.Expect.Subscription {
			t.Fatalf("Expected %q but got %q for Subscription", v.Expect.Subscription, actual.Subscription)
		}

		if actual.ResourceGroup != v.Expect.ResourceGroup {
			t.Fatalf("Expected %q but got %q for Resource Group", v.Expect.ResourceGroup, actual.ResourceGroup)
		}

		if actual.LogicAppWorkflowId.Name != v.Expect.LogicAppWorkflowId.Name {
			t.Fatalf("Expected %q but got %q for Workflow Name", v.Expect.LogicAppWorkflowId.Name, actual.LogicAppWorkflowId.Name)
		}

		if actual.Name != v.Expect.Name {
			t.Fatalf("Expected %q but got %q for Name", v.Expect.Name, actual.Name)
		}
	}
}
