package config

import "fmt"

const (
	awsMSKIAM = "aws_msk_iam"
)

type AuthType uint

const (
	NoAuth AuthType = iota
	AWSMSKIAM
)

func ValidateAuthValue(value string) (AuthType, error) {
	switch value {
	case "none":
		return NoAuth, nil
	case awsMSKIAM:
		return AWSMSKIAM, nil
	default:
		return NoAuth, fmt.Errorf("auth value is missing/incorrect; possible values: [none, %s]", awsMSKIAM)
	}
}
