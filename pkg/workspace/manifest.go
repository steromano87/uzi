package workspace

type Manifest struct {
	Name               string `default:"My Harkonnen workspace"`
	VersionConstraints string `default:"> 0.0.0"`

	Profile string `default:"default"`

	Scenario      map[string]any
	Configuration Configuration
}
