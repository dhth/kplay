package cmd

type KafkaAuthenticationType uint

const (
	NoAuth KafkaAuthenticationType = iota
	SaslIamAuth
)
