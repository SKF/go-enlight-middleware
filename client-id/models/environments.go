package models

import "fmt"

type Environment string
type Environments []Environment

type EnvironmentMask uint8

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

func (envs Environments) Mask() EnvironmentMask {
	var mask uint8

	if len(envs) == 0 {
		return EnvironmentMask(0b1111) //nolint:mnd
	}

	for _, env := range envs {
		switch env {
		case Sandbox:
			mask |= 0b0001 //nolint:mnd
		case Test:
			mask |= 0b0010 //nolint:mnd
		case Staging:
			mask |= 0b0100 //nolint:mnd
		case Prod:
			mask |= 0b1000 //nolint:mnd
		}
	}

	return EnvironmentMask(mask)
}

func (mask EnvironmentMask) Disjoint(other EnvironmentMask) bool {
	return uint8(mask)&uint8(other) == 0
}
