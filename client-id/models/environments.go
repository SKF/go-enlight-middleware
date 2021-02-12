package models

import "fmt"

type Environment string
type Environments []Environment

const (
	Sandbox Environment = "sandbox"
	Test    Environment = "test"
	Staging Environment = "staging"
	Prod    Environment = "prod"
)

var AllEnvironments = Environments{
	Sandbox, Test, Staging, Prod,
}

func (e Environment) Validate() error {
	if !AllEnvironments.Contains(e) {
		return fmt.Errorf("`%s` must be one of %s", e, AllEnvironments)
	}

	return nil
}

func (envs Environments) Contains(e Environment) bool {
	for _, env := range envs {
		if env == e {
			return true
		}
	}

	return len(envs) == 0 && e.Validate() == nil
}
