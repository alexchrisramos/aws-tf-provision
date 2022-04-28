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
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// logCmd represents the log command
var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Display the logs for a pod",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Aliases: []string{"logs"},
	Args:    cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("log called")

		bundleRootDir := getBundleRootDir()
		podLogsDir := getPodLogsDir(bundleRootDir)

		// TODO: Make this a true cobra nested subcommand instead
		if len(args) == 0 || args[0] == "ls" || args[0] == "list" {
			listPodLogs(podLogsDir)
		} else {
			viewPodLog(args, podLogsDir)
		}
	},
}

func getPodLogsDir(bundleRootDir string) string {
	var podLogsDir string

	err := filepath.Walk(bundleRootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("Failure accessing path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() && info.Name() == "pods_logs" {
			podLogsDir = path
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking the path %q: %v\n", bundleRootDir, err)
	}

	if podLogsDir == "" {
		log.Fatalf("Failed to find pod logs dir within bundle directory: %s\n", bundleRootDir)
	}

	return podLogsDir
}

func listPodLogs(podLogsDir string) {
	var files []string
	// Find yaml files in directory -- Maybe clean this up a bit in the future
	err := filepath.Walk(podLogsDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".log" {
			files = append(files, path)
			return nil
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	podList := [][]string{}
	for _, file := range files {
		podMetadata := strings.Split(strings.TrimSuffix(file[strings.LastIndex(file, "/")+1:], ".log"), "_")
		podNamespace := podMetadata[0]
		podName := podMetadata[1]
		podList = append(podList, []string{podNamespace, podName})

		// fmt.Printf("Pod log file name: %v\n", file)
		// fmt.Printf("Pod namespace: %v\n", podNamespace)
		// fmt.Printf("Pod name: %v\n", podName)

	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Namespace", "Name"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)
	table.AppendBulk(podList) // Add Bulk Data
	table.Render()
}

func viewPodLog(args []string, podLogsDir string) {
	var files []string
	// Find yaml files in directory -- Maybe clean this up a bit in the future
	err := filepath.Walk(podLogsDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".log" {
			files = append(files, path)
			return nil
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	var podLogFile string
	for _, file := range files {
		filePodMetadata := strings.Split(strings.TrimSuffix(file[strings.LastIndex(file, "/")+1:], ".log"), "_")
		// podNamespace := filePodMetadata[0]
		// podName := filePodMetadata[1]

		if args[0] == filePodMetadata[0] && args[1] == filePodMetadata[1] {
			// fmt.Printf("Pod match found!\n")
			// fmt.Printf("Pod log file name: %v\n", file)
			// fmt.Printf("Pod namespace: %v\n", filePodMetadata[0])
			// fmt.Printf("Pod name: %v\n", filePodMetadata[1])
			podLogFile = file
		}
	}

	if podLogFile != "" {
		pager := os.Getenv("PAGER")
		if pager == "" {
			pager = "less"
		}

		cmd := exec.Command(pager, podLogFile)
		cmd.Stdout = os.Stdout

		fmt.Printf("Opening pod log file: %v", podLogFile)
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Could not open %v in `%v`: %v", podLogFile, pager, err)
		}
	} else {
		log.Fatalf("Could not find log file for %v pod in %v namespace.", args[1], args[0])
	}
}

func init() {
	rootCmd.AddCommand(logCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// logCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// logCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
