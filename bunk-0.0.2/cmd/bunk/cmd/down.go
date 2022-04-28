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
	"log"
	"os"
	"os/exec"
	"os/user"
	"runtime"

	"github.com/spf13/cobra"
)

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Destroy a kbk cluster for a bundle",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		down()
	},
}

func deleteKubernetesCluster() {
	cmd := exec.Command("k3d",
		"delete",
		"--all",
	)
	err := cmd.Run()
	if err != nil {
		log.Printf("Failed to remove k3d cluster: %s\n", err)
	} else {
		log.Printf("Successfully removed k3d cluster!\n")
	}
}

func deleteResourceDir(resourceDir string) {
	if _, err := os.Stat(resourceDir); os.IsNotExist(err) {
		log.Printf("Failed to remove resource directory: %s\n", err)
	} else {
		whoami, err := user.Current()
		if err != nil {
			log.Fatalf("Failed to get current user: %s", err)
		}

		// Fix directory permissions on linux
		hostOS := runtime.GOOS
		if _, err := os.Stat(resourceDir + "/db"); !os.IsNotExist(err) && hostOS == "linux" {
			cmd := exec.Command("/bin/sh", "-c", "sudo chown -R "+whoami.Username+" "+resourceDir+"/db")
			if err := cmd.Run(); err != nil {
				log.Fatalf("Failed to chown directory to user %s: %s", whoami.Username, err)
			}
		}

		if err := os.RemoveAll(resourceDir); err != nil {
			log.Fatal(err)
		}
		log.Printf("Successfully removed resource directory!\n")
	}
}

func down() {
	bundleRootDir := getBundleRootDir()
	resourceDir := bundleRootDir + "/.kbk"

	deleteKubernetesCluster()
	deleteResourceDir(resourceDir)
}

func init() {
	rootCmd.AddCommand(downCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
