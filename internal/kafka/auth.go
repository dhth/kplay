package kafka

type AuthType uint

const (
	NoAuth AuthType = iota
	SaslIamAuth
)
