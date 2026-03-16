package models

type Repository struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type Release struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Chart     string `yaml:"chart"`
	Version   string `yaml:"version"`
}

type SubHelmfileEntry struct {
	Path               string   `yaml:"path"`
	Selectors          []string `yaml:"selectors"`
	SelectorsInherited bool     `yaml:"selectorsInherited"`
}

type Helmfile struct {
	Repositories []Repository `yaml:"repositories"`
	Releases     []Release    `yaml:"releases"`
	Helmfiles    []any        `yaml:"helmfiles"`
}
