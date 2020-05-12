package parse

import (
	"fmt"
	"regexp"
)

type LogicAppTriggerId struct {
	LogicAppWorkflowId
	Name string
}

func LogicAppTriggerID(input string) (*LogicAppTriggerId, error) {
	// Format: <LogicAppWorkflowID>/triggers/<name>
	groups := regexp.MustCompile(`^(.+)/triggers/([^/]+)$`).FindStringSubmatch(input)
	if len(groups) != 3 {
		return nil, fmt.Errorf("faield to parse Logic App Trigger ID: %q", input)
	}

	workflow, name := groups[1], groups[2]
	workflowId, err := LogicAppWorkflowID(workflow)
	if err != nil {
		return nil, fmt.Errorf("parsing Workflow part of Logic App Trigger ID %q: %+v", input, err)
	}
	return &LogicAppTriggerId{
		*workflowId,
		name,
	}, nil
}
