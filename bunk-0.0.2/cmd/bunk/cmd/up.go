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
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	// Add sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// KubernetesItems : Structure of Kubernetes yaml files -- make this more clear/redefine
type KubernetesItems struct {
	Items []interface{} `yaml:"items"`
}

// KubernetesObject : Generic minimal structure for Kubernetes objects
type KubernetesObject struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace,omitempty"`
	} `yaml:"metadata"`
}

// DatabaseObject : Structure of each item in the k3d SQLite3 backend database
// type DatabaseObject struct {
// 	id             int
// 	name           string
// 	created        int
// 	deleted        int
// 	createRevision int
// 	prevRevision   int
// 	lease          int
// 	value          string
// 	oldValue       string
// }

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Create a kbk cluster for a bundle",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// log.Println("up called")
		up()
	},
}

func getBundleRootDir() string {
	var bundleRootDir string
	var cutString string

	// Set bundle dir; prio env var > cwd path
	// TODO: Reevaluate this priority
	if os.Getenv("BUNK_BUNDLE_DIR") != "" {
		bundleRootDir = os.Getenv("BUNK_BUNDLE_DIR")
	} else {
		workDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get work dir: %s\n", err)
		}

		// Split workdir path to find dir prefixed with 'bundle-'
		s := strings.SplitAfter(workDir, "/")
		for i := len(s) - 1; i >= 0; i-- {
			if strings.HasPrefix(s[i], "bundle-") == true {
				bundleRootDir = strings.TrimRight(workDir, cutString)
				break
			} else if s[i] != "/" {
				cutString += s[i]
			} else {
				log.Fatalf("Failed to find bundle root in work dir path: %s\n", workDir)
			}
		}
	}
	bundleRootDir = strings.TrimSuffix(bundleRootDir, "/")

	return bundleRootDir
}

func getAPIResourcesDir(bundleRootDir string) string {
	var apiResourcesDir string

	err := filepath.Walk(bundleRootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf("Failure accessing path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() && info.Name() == "api-resources" {
			apiResourcesDir = path
		}
		// log.Printf("visited %s file or dir: %q\n", info.Name(), path)
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking the path %q: %v\n", bundleRootDir, err)
	}

	if apiResourcesDir == "" {
		log.Fatalf("Failed to find api-resources dir within bundle directory: %s\n", bundleRootDir)
	}

	return apiResourcesDir
}

// func preflight() (string, string) {

// 	bundleRootDir := getBundleRootDir()
// 	apiResourcesDir := getAPIResourcesDir(bundleRootDir)

// 	log.Printf("api-resources dir: %s\n", apiResourcesDir)

// 	return bundleRootDir, apiResourceDir
// }

func initConfigDir(bundleRootDir string) string {
	err := os.Mkdir(bundleRootDir+"/.kbk", 0774)
	if err != nil {
		log.Fatalf("Failed to create .kbk directory at %s: %s\n", bundleRootDir, err)
	}
	resourceDir := bundleRootDir + "/.kbk"
	return resourceDir
}

func writeKubernetesResources(bundleRootDir string, apiResourcesDir string, resourceDir string) string {
	// Set sql file for writing
	kubernetesResourcesSQL := resourceDir + "/kubernetesResources.sql"
	// Set base resourceID to something out of range
	resourceID := 5000

	var files []string
	// Find yaml files in directory -- Maybe clean this up a bit in the future
	err := filepath.Walk(apiResourcesDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".yaml" {
			files = append(files, path)
			return nil
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Kill the file because we're lazy atm :)
	// TODO: remove this
	os.Remove(kubernetesResourcesSQL)

	// For each yaml file in apiresources dir
	for _, file := range files {
		// Get the file basename from its full path
		basename := file[strings.LastIndex(file, "/")+1:]

		// log.Println("DEBUG BASENAME: %s", basename)

		// Get api resource group from file name
		// TODO: fix this to better account for x.yaml files (e.g., pods.yaml)
		apiResourceGroup := strings.TrimSuffix(strings.SplitN(basename, ".", 2)[1], ".yaml")

		// Write out api resource groups that break path in database
		switch apiResourceGroup {
		case "yaml",
			"apps",
			"certificates.k8s.io",
			"coordination.k8s.io",
			"extensions",
			"networking.k8s.io",
			"rbac.authorization.k8s.io",
			"scheduling.k8s.io",
			"storage.k8s.io",
			"snapshot.storage.k8s.io":
			apiResourceGroup = ""
		}

		// log.Println("DEBUG APIRESOURCEGROUP: %s", apiResourceGroup)

		// Get api resource name from file name
		apiResourceName := strings.SplitN(basename, ".", 2)[0]

		// Modify api resource names for known inconsistencies
		switch apiResourceName {
		case "nodes":
			apiResourceName = "minions"
		case "endpoints":
			apiResourceName = "services/endpoints"
		case "services":
			apiResourceName = "services/specs"
		case "leases":
			apiResourceName = "leases/kube-node-lease"
		case "ingresses":
			apiResourceName = "ingress"
		case "podsecuritypolicies":
			apiResourceName = "podsecuritypolicy"
		}

		// Ignore secrets file, as it is not valid yaml
		if apiResourceName == "secrets" {
			continue
		} else {
			// Begin parsing yaml to create sql statements
			// Open file as read-only
			f, err := os.OpenFile(file, os.O_RDONLY, 0444)
			if err != nil {
				log.Fatal(err)
			}

			// Close file when finished
			defer f.Close()

			// Read the file
			content, err := ioutil.ReadAll(f)
			if err != nil {
				log.Fatal(err)
			}

			// Unmarshal yaml to json
			var kubernetesItems KubernetesItems
			if err := yaml.Unmarshal(content, &kubernetesItems); err != nil {
				log.Fatal(err)
			}

			// Pretty colors rock!
			green := color.New(color.FgGreen).PrintfFunc()
			yellow := color.New(color.FgYellow).PrintfFunc()
			// red := color.New(color.FgRed).PrintFunc()

			// Give the people some nice output
			if len(kubernetesItems.Items) == 0 {
				yellow("Skipping empty %s resource file: %s\n", apiResourceName, basename)
			} else {
				green("Writing %d %s resources from file: %s\n", len(kubernetesItems.Items), apiResourceName, basename)
			}

			// For each item in the Kubernetes resource file
			for i := 0; i < len(kubernetesItems.Items); i++ {
				// Marshal the json items to individual []byte
				// objectJSON, err := json.Marshal(kubernetesItems.Items[i])
				// if err != nil {
				// 	log.Fatal(err)
				// }

				// enc := json.NewEncoder(os.Stdout)
				// enc.SetEscapeHTML(false)
				// err := enc.Encode(objectJSON)
				// if err != nil {
				// 	log.Fatal(err)
				// }

				buffer := &bytes.Buffer{}
				enc := json.NewEncoder(buffer)
				enc.SetEscapeHTML(false)
				err := enc.Encode(kubernetesItems.Items[i])
				if err != nil {
					log.Fatal(err)
				}

				objectJSON := buffer.Bytes()

				// Preserve the raw JSON []byte and trim newline for printing >:(
				objectState := []byte(strings.TrimSuffix(string(objectJSON), "\n"))

				var kubernetesObject KubernetesObject

				// Unmarshal json to kubernetes object struct to write as sql statements
				if err := json.Unmarshal(objectJSON, &kubernetesObject); err != nil {
					log.Fatal(err)
				}

				// Print the full object json
				// log.Printf("%s\n\n", objectState)

				var sqlInsertStatement string = ""
				// var apiResourceName string
				// var apiResourceNamespaced bool
				// var objectState []byte
				// var objectName string
				// var objectNamespace string
				objectName := kubernetesObject.Metadata.Name
				objectNamespace := kubernetesObject.Metadata.Namespace

				sqlInsertStatement = fmt.Sprintf(
					"INSERT INTO kine(id, name, created, deleted, create_revision, prev_revision, lease, value, old_value) "+
						"VALUES(%d, '/registry/",
					resourceID)

				// Add api resource group to path if it exists
				if apiResourceGroup != "" {
					sqlInsertStatement += fmt.Sprintf("%s/", apiResourceGroup)
				}

				// Add api resource name
				sqlInsertStatement += fmt.Sprintf("%s/", apiResourceName)

				// Add namespace if it exists
				if objectNamespace != "" {
					sqlInsertStatement += fmt.Sprintf("%s/", objectNamespace)
				}

				// Escape objectState s/\'/\'\'/g as string and convert back to byte
				objectState = []byte(strings.ReplaceAll(string(objectState), "'", "''"))

				// Add required additional fields
				sqlInsertStatement += fmt.Sprintf("%s', 1, 0, %d, %d, 0, '%s', '%s');", objectName, resourceID+1, resourceID+2, objectState, objectState)

				// Increment resourceID
				// TODO: I think 3 is sufficient -- need to test this
				resourceID += 4

				// Open file; write the file; close the file; if closing file fails, then sad panda
				sqlFile, err := os.OpenFile(kubernetesResourcesSQL, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
				if err != nil {
					log.Fatal(err)
				}
				if _, err := sqlFile.Write([]byte(sqlInsertStatement + "\n")); err != nil {
					sqlFile.Close() // ignore error; Write error takes precedence
					log.Fatal(err)
				}

				// Close the file when done
				// This must be explicit or we will hit too many open files
				if err := sqlFile.Close(); err != nil {
					log.Fatalf("Failed to close sql file: %v", err)
				}
			}
		}
	}

	return kubernetesResourcesSQL
}

func createKubernetesCluster(kubernetesResourcesSQL string, resourceDir string) {
	log.Println("Creating k3d cluster")
	cmd := exec.Command("k3d",
		"create",
		"--workers", "0",
		"--volume", resourceDir+"/db:/var/lib/rancher/k3s/server/db/",
		"--server-arg", "--disable-agent",
		"--server-arg", "--no-deploy=coredns",
		"--server-arg", "--no-deploy=servicelb",
		"--server-arg", "--no-deploy=traefik",
		"--server-arg", "--no-deploy=local-storage",
		"--server-arg", "--no-deploy=metrics-server",
		"--server-arg", "--kube-apiserver-arg=event-ttl=168h0m0s",
		"--server-arg", "--kube-controller-arg=disable-attach-detach-reconcile-sync",
		"--server-arg", "--kube-controller-arg=controllers=-attachdetach,-clusterrole-aggregation,-cronjob,-csrapproving,-csrcleaner,-csrsigning,-daemonset,-deployment,-disruption,-endpoint,-garbagecollector,-horizontalpodautoscaling,-job,-namespace,-nodeipam,-nodelifecycle,-persistentvolume-binder,-persistentvolume-expander,-podgc,-pv-protection,-pvc-protection,-replicaset,-replicationcontroller,-resourcequota,-root-ca-cert-publisher,-serviceaccount,-serviceaccount-token,-statefulset,-ttl",
		"--server-arg", "--disable-scheduler",
		"--server-arg", "--disable-cloud-controller",
		"--server-arg", "--disable-network-policy",
		"--server-arg", "--no-flannel",
		"--wait", "60",
	)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	log.Println("Stopping k3d cluster")
	cmd = exec.Command("k3d",
		"stop",
		"--all",
	)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	whoami, err := user.Current()
	if err != nil {
		log.Fatalf("Failed to get current user: %s", err)
	}

	// Fix directory permissions on linux
	hostOS := runtime.GOOS
	if hostOS == "linux" {
		cmd = exec.Command("/bin/sh", "-c", "sudo chown -R "+whoami.Username+" "+resourceDir+"/db")
		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to chown directory to user %s: %s", whoami.Username, err)
		}
	}

	log.Println("Adding cluster resources")

	database, err := sql.Open("sqlite3", resourceDir+"/db/state.db")
	if err != nil {
		log.Fatal(err)
	}

	defer database.Close()

	kubernetesSQLBackendData, err := ioutil.ReadFile(kubernetesResourcesSQL)
	if err != nil {
		log.Fatalf("Failed opening file: %s", err)
	}

	if result, err := database.Exec(string(kubernetesSQLBackendData)); err != nil {
		log.Fatalf("Error executing query %v: %v", result, err)
	}

	// Fix directory permissions on linux
	if hostOS == "linux" {
		cmd = exec.Command("/bin/sh", "-c", "sudo chown -R root "+resourceDir+"/db")
		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to chown directory to user root: %s", err)
		}
	}

	log.Println("Starting k3d cluster")
	cmd = exec.Command("k3d",
		"start",
		"--all",
	)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	log.Printf("k3d cluster created! Please access the cluster with:\nexport KUBECONFIG=\"$(k3d get-kubeconfig --name='k3s-default')\"\n")
}

func up() {
	bundleRootDir := getBundleRootDir()
	apiResourcesDir := getAPIResourcesDir(bundleRootDir)
	resourceDir := initConfigDir(bundleRootDir)

	log.Printf("Bundle root dir: %s\n", bundleRootDir)
	log.Printf("api-resources dir: %s\n", apiResourcesDir)

	kubernetesResourcesSQL := writeKubernetesResources(bundleRootDir, apiResourcesDir, resourceDir)

	createKubernetesCluster(kubernetesResourcesSQL, resourceDir)
}

func init() {
	rootCmd.AddCommand(upCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// upCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// upCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
