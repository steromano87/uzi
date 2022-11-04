package configuration

type Load struct {
	Script        string
	MaxIterations uint64
	LoadProfiles  []LoadProfile
}

type LoadProfile struct {
	Type string
	Spec map[string]any
}
