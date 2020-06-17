/**
 * Copyright 2020 Appvia Ltd <info@appvia.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/appvia/kube-devx/pkg/kev/bootstrap"
	"github.com/disiqueira/gotree"
	"github.com/spf13/cobra"
)

var initLongDesc = `(init) reuses one or more docker-compose files to initialise a cloud native app.

Examples:

  # Initialise an app definition with a single docker-compose file
  $ kev init -c docker-compose.yaml

  # Initialise an app definition with multiple docker-compose files.
  # These will be interpreted as one file.
  $ kev init -c docker-compose.yaml -c docker-compose.other.yaml

  # Initialise an app definition with a deployment environment.
  $ kev init -e staging -c docker-compose.yaml

  # Initialise an app definition with multiple deployment environments.
  $ kev init -e staging -e dev -e prod -c docker-compose.yaml`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Reuses project docker-compose file(s) to initialise an app definition.",
	Long:  initLongDesc,
	RunE:  runInitCmd,
}

func init() {
	flags := initCmd.Flags()
	flags.SortFlags = false

	flags.StringSliceP(
		"compose-file",
		"c",
		[]string{},
		"Compose file to use as application base - use multiple flags for additional files",
	)
	if err := initCmd.MarkFlagRequired("compose-file"); err != nil {
		log.Fatal(err)
	}

	flags.StringSliceP(
		"environment",
		"e",
		[]string{},
		"Deployment environments in addition to application base (optional) ",
	)

	rootCmd.AddCommand(initCmd)
}

func runInitCmd(cmd *cobra.Command, _ []string) error {
	composeFiles, _ := cmd.Flags().GetStringSlice("compose-file")
	envs, _ := cmd.Flags().GetStringSlice("environment")

	def, err := bootstrap.NewApp(BaseDir, composeFiles, envs)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(BaseDir, os.ModePerm); err != nil {
		return err
	}

	if err := ioutil.WriteFile(def.BaseCompose.File, def.BaseCompose.Content, os.ModePerm); err != nil {
		return err
	}

	if err := ioutil.WriteFile(def.Config.File, def.Config.Content, os.ModePerm); err != nil {
		return err
	}

	fmt.Printf("🚀 App initialised")

	defSource := gotree.New("\n\nSource compose file(s)")
	for _, e := range composeFiles {
		defSource.Add(e)
	}
	fmt.Println(defSource.Print())

	defTree := gotree.New("\n\nApplication configuration files")
	defTree.Add(def.BaseCompose.File)
	defTree.Add(def.Config.File)

	for _, env := range def.Envs {
		if err := os.MkdirAll(env.Dir(), os.ModePerm); err != nil {
			return err
		}

		if err := ioutil.WriteFile(env.File, env.Content, os.ModePerm); err != nil {
			return err
		}

		defTree.Add(env.File)
	}

	fmt.Println(defTree.Print())

	return nil
}
