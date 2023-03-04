package configuration

type Injector struct {
	Description string
	Address     string
	Local       bool
	Optional    bool
	Weight      uint
}
