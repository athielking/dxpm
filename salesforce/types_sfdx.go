package salesforce

//SfdxProject represents the sfdx-project.json file in sfdx project root directory.
type SfdxProject struct {
	PackageDirectories []struct {
		Path          string                  `json:"path"`
		Default       bool                    `json:"default"`
		PackageName   string                  `json:"package"`
		VersionName   string                  `json:"versionName"`
		VersionNumber string                  `json:"versionNumber"`
		Dependencies  []SfdxProjectDependency `json:"dependencies"`
	} `json:"packageDirectories"`
	Namespace        string            `json:"namespace"`
	SfdcLoginURL     string            `json:"sfdcLoginUrl"`
	SourceAPIVersion string            `json:"sourceApiVersion"`
	PackageAliases   map[string]string `json:"packageAliases"`
}

//SfdxProjectDependency represents a dependent package for this project
type SfdxProjectDependency struct {
	PackageName string `json:"package"`
}
