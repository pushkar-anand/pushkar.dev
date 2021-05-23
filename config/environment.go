package config

type Environment string

const (
	EnvProduction  Environment = "PRODUCTION"
	EnvDevelopment Environment = "DEVELOPMENT"
)

func (e Environment) String() string {
	return string(e)
}
