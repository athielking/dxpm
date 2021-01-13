package salesforce

// Org represents a Salesforce non-scratch org
type Org struct {
	UserName      string
	OrgID         string `json:"OrgId"`
	AccessToken   string
	IsDevHub      bool
	Alias         string
	DefaultMarker string
}

// ScratchOrg represents a Salesforce scratch org
type ScratchOrg struct {
	UserName       string
	OrgID          string `json:"OrgId"`
	DevHubOrgID    string `json:"DevHubOrgId"`
	AccessToken    string
	Alias          string
	Status         string
	IsExpired      bool
	ExpirationDate string
}

type orgListResponse struct {
	Status int
	Result struct {
		NonScratchOrgs []Org
		ScratchOrgs    []ScratchOrg
	}
}

