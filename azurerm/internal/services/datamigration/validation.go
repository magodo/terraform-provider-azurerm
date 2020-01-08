package datamigration

import (
	"fmt"
	"regexp"
)

func validateName(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}
	validName := regexp.MustCompile(`^[\d\w]+[\d\w\-_.]*$`)
	if !validName.MatchString(v) {
		return nil, []error{fmt.Errorf("invalid format of %q", k)}
	}
	return nil, nil
}

func validateTaskName(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}
	validName := regexp.MustCompile(`^[^\W_][\w-.]{1,61}$`)
	if !validName.MatchString(v) {
		return nil, []error{fmt.Errorf("inlivad format of %q", k)}
	}
	return nil, nil
}
