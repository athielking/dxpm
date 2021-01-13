package salesforce

// Pkg represents a Salesforce package object
type Pkg struct {
	Name string
	ID   string `json:"Id"`
}

// PkgVersion represents a Salesforce package version
type PkgVersion struct {
	Name        string `json:"Package2Name"`
	ID          string `json:"SubscriberPackageVersionId"`
	PackageID   string `json:"Package2Id"`
	VersionName string `json:"Name"`
	Version     string
}

//SubscriberPkgVersion represents a SubscriberPackageVersion object from the tooling api
type SubscriberPkgVersion struct {
	ID           string
	Name         string
	PackageID    string `json:"SubscriberPackageId"`
	MajorVersion int
	MinorVersion int
	PatchVersion int
	BuildNumber  int
	PackageType  string `json:"Package2ContainerOptions"`
	Dependencies struct {
		Ids []struct {
			SubscriberPackageVersionID string `json:"subscriberPackageVersionId"`
		} `json:"ids"`
	}
}

type subscriberPackageDependency struct {
	SubscriberPackageVersionID string `json:"subscriberPackageVersionId"`
}

type soqlSubscriberPkgVersion struct {
	Status int
	Result struct {
		Size           int
		EntityTypeName string
		Records        []SubscriberPkgVersion
	}
}

//SubscriberPkg represents a SubscriberPackage object from the tooling api
type SubscriberPkg struct {
	Name string
}

//InstalledPkg represents a response item from sfdx force:package:installed:list
type InstalledPkg struct {
	ID                         string `json:"Id"`
	SubscriberPackageID        string `json:"SubscriberPackageId"`
	SubscriberPackageName      string
	SubscriberPackageVersionID string `json:"SubscriberPackageVersionId"`
}

type installedPkgResponse struct {
	Status int
	Result []InstalledPkg
}

type pkgResponse struct {
	Status int
	Result []Pkg
}

type versionResponse struct {
	Status int
	Result []PkgVersion
}

type soqlResult struct {
	Status int
	Result struct {
		Size           int
		EntityTypeName string
		Records        []interface{}
	}
}
