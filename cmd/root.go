/*
Copyright 2021 Shiwen Cheng

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
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var Version = "unknown"

type flags struct {
	KubeConfig string
}

func defaultKubeConfig() string {
	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config")
	}
	return ""
}

func Execute() {
	cmd := NewCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func NewCommand() *cobra.Command {
	flags := &flags{}
	cmd := &cobra.Command{
		Use:           "kubectl-resource-versions",
		Short:         "Print the API resources with the supported API versions",
		Long:          "Print the API resources along with the supported API versions in the form of group/version on the server",
		Example:       "  # Print the API resources with the supported API versions\n  kubectl-resource-versions\n  # Print by kubectl plugin\n  kubectl resource-versions",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       Version,
		RunE: func(c *cobra.Command, args []string) error {
			return runE(flags)
		},
	}
	cmd.PersistentFlags().StringVar(&flags.KubeConfig, "kubeconfig", defaultKubeConfig(), "path to the kubeconfig file to use for CLI requests")
	return cmd
}

func runE(flags *flags) (err error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", flags.KubeConfig)
	if err != nil {
		return
	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return
	}

	// request /openapi/v2
	body, err := clientset.RESTClient().Get().AbsPath("/openapi/v2").Do(context.TODO()).Raw()
	if err != nil {
		return
	}
	var document struct {
		Paths map[string]interface{} `json:"paths,omitempty"`
	}
	err = json.Unmarshal(body, &document)
	if err != nil {
		return
	}

	printResult(document.Paths)
	return
}

func printResult(paths map[string]interface{}) {
	// filter valid paths
	validPaths := make([]string, 0)
	for path := range paths {
		if (strings.HasPrefix(path, "/api/") && !strings.HasSuffix(path, "/") && len(strings.Split(path, "/")) == 4) ||
			(strings.HasPrefix(path, "/apis/") && !strings.HasSuffix(path, "/") && len(strings.Split(path, "/")) == 5) {
			validPaths = append(validPaths, path)
		}
	}
	sort.Strings(validPaths)

	// valid paths -> resource map { resource: apiVersions }
	resources := make(map[string][]string)
	for _, path := range validPaths {
		items := strings.Split(path, "/")
		resource := items[len(items)-1]
		apiVersion := strings.Join(items[2:len(items)-1], "/")
		if _, ok := resources[resource]; !ok {
			resources[resource] = make([]string, 0)
		}
		resources[resource] = append(resources[resource], apiVersion)
	}

	// sort resources
	maxLen := 0
	resourceKeys := make([]string, 0)
	for resource := range resources {
		resourceKeys = append(resourceKeys, resource)
		if len(resource) > maxLen {
			maxLen = len(resource)
		}
	}
	sort.Strings(resourceKeys)

	// print resource versions
	format := "%-" + strconv.Itoa(maxLen+4) + "s%s\n"
	fmt.Printf(format, "RESOURCE", "API VERSIONS")
	for _, resource := range resourceKeys {
		fmt.Printf(format, resource, strings.Join(resources[resource], ", "))
	}
}
