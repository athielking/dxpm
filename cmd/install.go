/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"dxpm/salesforce"
)

var org string
var pkg string
var create bool
var filePath string
var saveDep bool

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install package and dependencies into a target org",
	Long: `Installs pagages and dependencies into a target org or scratch org.  ID's or Aliases
can be used to target orgs and packages.

Examples:

dxpm install -o <ORG ID or ALIAS> : Must be ran from within an SFDX Project and will attempt 
to install all project dependencies into the specified org

dxpm install -o <ORG ID or ALIAS> -p <PACKAGE NAME or ID>: Will install the specified package 
and all dependencies to the target org.

dxpm install -p <PACKAGE NAME or ID> -c -f <Path to scratch-def.json> : Will first 
create a scratch org with the specified alias and then install the package and dependencies`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(org) > 0 && len(pkg) < 0 {
			return salesforce.CheckSFDX()
		}

		if saveDep {
			return salesforce.CheckSFDX()
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		orgSet := len(org) > 0
		pkgSet := len(pkg) > 0

		if orgSet && pkgSet {
			err := salesforce.InstallPackage(org, pkg)
			if err != nil {
				fmt.Println(err)
			}

			return
		}

	},
}

func init() {
	installCmd.Flags().StringVarP(&org, "org", "o", "", "Org Alias or ID to install package to")
	installCmd.MarkFlagRequired("org")

	installCmd.Flags().StringVarP(&pkg, "pkg", "p", "", "Package Alias or ID to install")
	installCmd.Flags().BoolVarP(&create, "create", "c", false, "Creates a new scratch org from file")
	installCmd.Flags().StringVarP(&filePath, "file", "f", "", "Scratch Org Definition File Path")
	installCmd.Flags().BoolVarP(&saveDep, "save", "s", false, "Attempts to save package as a dependency to sfdx-project.json")

	rootCmd.AddCommand(installCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
