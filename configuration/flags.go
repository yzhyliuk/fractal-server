package configuration

var Mode string

const Dev = "dev"
const Prod = "prod"
const DebugProd = "debugProd"

func IsProduction() bool {
	return Mode == Prod
}

func IsDebugProd() bool {
	return Mode == DebugProd
}
