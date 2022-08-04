package configuration

var Mode string

const Dev = "dev"
const Prod = "prod"

func IsProduction() bool {
	return Mode == Prod
}
