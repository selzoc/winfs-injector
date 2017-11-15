package tile

type Metadata struct {
	Releases []Release
	Other    map[string]interface{} `yaml:",inline"`
}

type Release struct {
	Name    string
	File    string
	Version string
}
