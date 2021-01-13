package salesforce

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	defaultMarker      = "(D)"
	packagePrefix      = "0Ho"
	versionPrefix      = "04t"
	orgPrefix          = "00D"
	projectFileName    = "sfdx-project.json"
	managedPackageType = "Managed"
)

var orgs []Org
var scrOrgs []ScratchOrg
var projectPath string
var pkgVersions []PkgVersion
var installedPkgs []InstalledPkg

// CheckCli searches for the sfdx cli in the
// directories named by the PATH environment variable.
func CheckCli() error {

	_, err := exec.LookPath("sfdx")
	if err != nil {
		return errors.New("SFDX CLI not found on %PATH%")
	}

	return nil
}

// DevHub searches your sfdx orgs for the org marked
// as your default DevHub org.
func DevHub() (*Org, error) {
	if err := getOrgs(); err != nil {
		return nil, err
	}

	var devHub Org
	for _, org := range orgs {
		if org.IsDevHub && org.DefaultMarker == defaultMarker {
			devHub = org
			break
		}
	}

	if &devHub == nil {
		return nil, errors.New("No default dev hub org found")
	}

	return &devHub, nil
}

// CheckSFDX searches the current directory and parent directories
// for a valid Salesforce DX project.
func CheckSFDX() error {
	if len(projectPath) > 0 {
		return nil
	}

	fmt.Println("Locating SFDX Project File...")
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	projectPath, err = findSfdxProject(wd)
	if err != nil {
		return err
	}

	fmt.Println("Project File Found: " + projectPath)
	return nil
}

//InstallPackage installs the specified package to the specified org and updates dependencies in the project file
func InstallPackage(org string, pkg string) error {
	if err := CheckCli(); err != nil {
		return err
	}

	if err := CheckSFDX(); err != nil {
		return err
	}

	org, err := getOrgUserID(org)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(pkg, versionPrefix) {
		pkg, err = getPkgVersionID(pkg)

		if err != nil {
			return err
		}
	}

	err = InstallDependencies(org, pkg)
	if err != nil {
		return err
	}

	installed := isPkgInstalled(org, pkg)

	if !installed {
		err = sfdx("force:package:install", "--package", pkg, "-u", org, "-w 100")
		if err != nil {
			return err
		}
	}

	err = upsertDependencyToProjectFile(org, pkg)
	if err != nil {
		return err
	}

	return nil
}

//InstallDependencies finds the required dependencies and installs them prior to the target package
func InstallDependencies(org string, pkg string) error {

	mainPkg, err := getSubscriberPkgVersion(org, pkg)
	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("Installing Dependencies for package: %s - %s", mainPkg.Name, mainPkg.ID))
	for _, dep := range mainPkg.Dependencies.Ids {

		err = InstallPackage(org, dep.SubscriberPackageVersionID)

		if err != nil {
			return err
		}
	}

	return nil
}

//UninstallPackage uninstalls the specified package from the specified org and removes dependencies from the project file
func UninstallPackage(org string, pkg string) error {
	if err := CheckCli(); err != nil {
		return err
	}

	if err := CheckSFDX(); err != nil {
		return err
	}

	org, err := getOrgUserID(org)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(pkg, versionPrefix) {
		pkg, err = getPkgVersionID(pkg)

		if err != nil {
			return err
		}
	}

	err = sfdx("force:package:uninstall", "--package", pkg, "-u", org)
	if err != nil {
		return err
	}

	err = removeDependencyFromProjectFile(pkg)
	if err != nil {
		return err
	}

	return nil
}

func findSfdxProject(path string) (string, error) {

	s := []string{path, projectFileName}
	filePath := strings.Join(s, string(os.PathSeparator))

	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {

		parent := filepath.Dir(path)
		if parent == path {
			return "", errors.New("This directory does not contain a SFDX project")
		}

		return findSfdxProject(parent)
	}

	return filePath, nil
}

func getOrgs() error {

	if len(orgs) > 0 {
		return nil
	}

	if err := CheckCli(); err != nil {
		return err
	}

	jsonBytes, err := sfdxJ("force:org:list")
	if err != nil {
		return err
	}

	var resp orgListResponse
	err = json.Unmarshal(jsonBytes, &resp)

	if err != nil {
		return err
	}

	orgs = resp.Result.NonScratchOrgs
	scrOrgs = resp.Result.ScratchOrgs

	return nil
}

func getOrgUserID(alias string) (string, error) {

	//already in the username form
	if strings.Contains(alias, "@") {
		return alias, nil
	}

	if err := getOrgs(); err != nil {
		return "", err
	}

	isID := strings.HasPrefix(alias, orgPrefix)

	for _, org := range orgs {
		if isID && org.OrgID == alias {
			return org.UserName, nil
		}

		if !isID && org.Alias == alias {
			return org.UserName, nil
		}
	}

	for _, org := range scrOrgs {
		if isID && org.OrgID == alias {
			return org.UserName, nil
		}

		if !isID && org.Alias == alias {
			return org.UserName, nil
		}
	}

	return "", errors.New("Failed to locate org with alias: " + alias)
}

func getPkgVersionID(alias string) (string, error) {

	if err := getPkgVersions(); err != nil {
		return "", err
	}

	isID := strings.HasPrefix(alias, packagePrefix)
	version := "LATEST"
	versionNum := "0"
	versionID := ""

	if !isID && strings.Contains(alias, "@") {
		split := strings.Split("@", alias)
		version = split[1]
	}

	for _, ver := range pkgVersions {
		if isID && ver.PackageID == alias {
			return ver.ID, nil
		}

		if !isID && ver.Name == alias {
			if version != "LATEST" && ver.Version == version {
				return ver.ID, nil
			}

			//Find the latest version
			// If ver.Version > versionNum
			if strings.Compare(ver.Version, versionNum) == 1 {
				versionNum = ver.Version
				versionID = ver.ID
			}
		}
	}

	if versionID != "" {
		return versionID, nil
	}

	return "", errors.New("Failed to locate package version with alias: " + alias)

}

func getPkgVersion(ID string) (*PkgVersion, error) {
	if err := getPkgVersions(); err != nil {
		return nil, err
	}

	for _, ver := range pkgVersions {
		if ver.ID == ID {
			return &ver, nil
		}
	}

	return nil, errors.New("Failed to locate package version: " + ID)
}

func getPkgVersions() error {
	if len(pkgVersions) > 0 {
		return nil
	}

	if err := CheckCli(); err != nil {
		return err
	}

	jsonBytes, err := sfdxJ("force:package:version:list")
	if err != nil {
		return err
	}

	var resp versionResponse
	err = json.Unmarshal(jsonBytes, &resp)

	if err != nil {
		return err
	}

	pkgVersions = resp.Result

	return nil
}

func getSubscriberPkgVersion(org string, ID string) (*SubscriberPkgVersion, error) {
	soql := fmt.Sprintf("SELECT Id, SubscriberPackageId, MajorVersion, MinorVersion, PatchVersion, BuildNumber, Package2ContainerOptions, Dependencies FROM SubscriberPackageVersion WHERE Id='%s'", ID)

	jsonBytes, err := sfdxJ("force:data:soql:query", "-u", org, "-t", "-q", soql)
	if err != nil {
		return nil, err
	}

	var response soqlSubscriberPkgVersion
	err = json.Unmarshal(jsonBytes, &response)
	if err != nil {
		return nil, err
	}

	if response.Result.Size != 1 {
		panic(fmt.Errorf("More than 1 Subscriber Package Version with ID: %s", ID))
	}

	pkv := response.Result.Records[0]

	pkg, err := getSubscriberPkg(org, pkv.PackageID)
	if err != nil {
		return nil, err
	}

	pkv.Name = pkg.Name

	return &pkv, nil
}

func getSubscriberPkg(org string, ID string) (*SubscriberPkg, error) {
	soql := fmt.Sprintf("SELECT Name FROM SubscriberPackage WHERE Id='%s'", ID)
	jsonBytes, err := sfdxJ("force:data:soql:query", "-u", org, "-t", "-q", soql)
	if err != nil {
		return nil, err
	}

	var response soqlResult
	err = json.Unmarshal(jsonBytes, &response)
	if err != nil {
		return nil, err
	}

	if response.Result.Size != 1 {
		panic(fmt.Errorf("More than 1 Subscriber Package with ID: %s", ID))
	}

	jsonBytes, err = json.Marshal(response.Result.Records[0])
	if err != nil {
		return nil, err
	}

	var pkg SubscriberPkg
	err = json.Unmarshal(jsonBytes, &pkg)
	if err != nil {
		return nil, err
	}

	return &pkg, nil
}

func getInstalledPackages(org string) error {
	if len(installedPkgs) > 0 {
		return nil
	}

	if err := CheckCli(); err != nil {
		return err
	}

	jsonBytes, err := sfdxJ("force:package:installed:list", "-u", org)
	if err != nil {
		return err
	}

	var resp installedPkgResponse
	err = json.Unmarshal(jsonBytes, &resp)

	if err != nil {
		return err
	}

	installedPkgs = resp.Result

	return nil
}

func isPkgInstalled(org string, pkgVersionID string) bool {
	if err := getInstalledPackages(org); err != nil {
		panic(err)
	}

	for _, pkg := range installedPkgs {
		if pkg.SubscriberPackageVersionID == pkgVersionID {
			return true
		}
	}

	return false
}

func upsertDependencyToProjectFile(org string, pkgVersionID string) error {

	data, err := ioutil.ReadFile(projectPath)
	if err != nil {
		return err
	}

	var proj SfdxProject
	err = json.Unmarshal(data, &proj)
	if err != nil {
		return err
	}

	pkgVersion, err := getSubscriberPkgVersion(org, pkgVersionID)
	if err != nil {
		return err
	}

	var pkgDep *SfdxProjectDependency
	for _, dep := range proj.PackageDirectories[0].Dependencies {
		if dep.PackageName == pkgVersion.Name {
			pkgDep = &dep
			break
		}
	}

	//Only add the dependency if it does not exist
	if pkgDep == nil {
		pkgDep = &SfdxProjectDependency{
			PackageName: pkgVersion.Name,
		}

		proj.PackageDirectories[0].Dependencies = append(proj.PackageDirectories[0].Dependencies, *pkgDep)
	}

	// We could set unmanaged package versions in the dependency but do we want to?
	// I envision we could do a dxpm update which would update dependent packages to the latest version anyways

	// if pkgVersion.PackageType != managedPackageType {
	// 	pkgDep.VersionNumber = fmt.Sprintf("%d.%d.%d.%d", pkgVersion.MajorVersion, pkgVersion.MinorVersion, pkgVersion.PatchVersion, pkgVersion.BuildNumber)
	// 	proj.PackageAliases[pkgDep.PackageName] = pkgVersion.PackageID
	// }

	proj.PackageAliases[pkgDep.PackageName] = pkgVersionID

	bytes, err := json.MarshalIndent(proj, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(projectPath, bytes, 0777)
	if err != nil {
		return err
	}

	return nil
}

func removeDependencyFromProjectFile(pkgVersionID string) error {

	data, err := ioutil.ReadFile(projectPath)
	if err != nil {
		return err
	}

	var proj SfdxProject
	err = json.Unmarshal(data, &proj)
	if err != nil {
		return err
	}

	pkgVersion, err := getPkgVersion(pkgVersionID)
	if err != nil {
		return err
	}

	pkgDep := SfdxProjectDependency{
		PackageName: pkgVersion.Name,
	}

	proj.PackageDirectories[0].Dependencies = make([]SfdxProjectDependency, 0, len(proj.PackageDirectories[0].Dependencies))
	for _, ver := range proj.PackageDirectories[0].Dependencies {
		if ver.PackageName == pkgDep.PackageName {
			continue
		}

		proj.PackageDirectories[0].Dependencies = append(proj.PackageDirectories[0].Dependencies, ver)
	}
	delete(proj.PackageAliases, pkgDep.PackageName)

	bytes, err := json.MarshalIndent(proj, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(projectPath, bytes, 0777)
	if err != nil {
		return err
	}

	return nil
}

//sfdx run sfdx command with os.Stdout
func sfdx(arg ...string) error {
	sfdx := exec.Command("sfdx", arg...)
	sfdx.Stdout = os.Stdout
	err := sfdx.Run()

	return err
}

//sfdxJ run sfdx command with JSON output
func sfdxJ(arg ...string) ([]byte, error) {

	arg = append(arg, "--json")

	sfdx := exec.Command("sfdx", arg...)
	buf := new(bytes.Buffer)
	sfdx.Stdout = buf

	err := sfdx.Run()

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
