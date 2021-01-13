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
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"dxpm/salesforce"
)

var devHub bool
var id string

// orgCmd represents the org command
var orgCmd = &cobra.Command{
	Use:   "org",
	Short: "org retrieves information about your registered sfdx orgs",
	Long: `org command retrieves information about the Salesforce orgs 
	registered with sfdx.  Both Scratch and Non-Scratch orgs
	
	Examples:
	
	dxpm org --dev : retrieves information about your DevHub org`,
	Args: func(cmd *cobra.Command, args []string) error {

		if len(id) < 1 && !devHub {
			return errors.New("At least one flag must be specified")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		if devHub {
			dev, err := salesforce.DevHub()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("Org ID:   %s\n", dev.OrgID)
			fmt.Printf("UserName: %s\n", dev.UserName)
		}

	},
}

func init() {
	orgCmd.Flags().BoolVarP(&devHub, "dev", "d", false, "Find default DevHub")
	orgCmd.Flags().StringVarP(&id, "id", "i", "", "Find org by ID")

	rootCmd.AddCommand(orgCmd)
}
