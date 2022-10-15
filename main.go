/*
Copyright 2016 The Kubernetes Authors.
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

// Note: the example only works with the code within the same release/branch.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/jedib0t/go-pretty/v6/table"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

var (
	clientSet  *kubernetes.Clientset
	kubeconfig *string
)

func init() {
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
}

func switchContext(context string) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
}

func buildConfigFromFlags(context, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

func getConfigFromConfig(context, kubeconfigPath string) string {
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{})
	clientConfig, err := config.RawConfig()
	if err != nil {
		panic(err.Error())
	}
	clusterRegex := fmt.Sprintf("%s.*", context)
	for n := range clientConfig.Contexts {
		re := regexp.MustCompile(clusterRegex)
		result := re.MatchString(n)
		if result {
			return re.FindString(n)
		}
	}
	return ""
}

type ContainerList struct {
	deployment map[string]ContainerInfo
}

type ContainerInfo struct {
	Name      string
	Namespace string
}

func getDeployment() ContainerList {
	cl := ContainerList{
		deployment: map[string]ContainerInfo{},
	}

	namespaces := []string{
		"kube-system",
	}
	for _, ns := range namespaces {
		resource, err := clientSet.AppsV1().Deployments(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, d := range resource.Items {
			for _, c := range d.Spec.Template.Spec.Containers {
				imageName := c.Image
				containerName := c.Name
				m := regexp.MustCompile("^registry.+net/")
				cl.deployment[containerName] = ContainerInfo{
					Name:      m.ReplaceAllString(imageName, ""),
					Namespace: ns,
				}
			}
		}
	}

	return cl
}

func compareComponents(c1 string, c2 string, n string) {
	var l1 ContainerList
	var l2 ContainerList

	var currentcontext string
	switch n {
	case "deployment":
		currentcontext = getConfigFromConfig(c1, *kubeconfig)
		switchContext(currentcontext)
		l1 = getDeployment()
		currentcontext = getConfigFromConfig(c2, *kubeconfig)
		switchContext(currentcontext)
		l2 = getDeployment()
	}

	set := make(map[string]bool)

	for i := range l1.deployment {
		if !set[i] {
			set[i] = true
		}
	}

	for i := range l2.deployment {
		if !set[i] {
			set[i] = true
		}
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Resource", "Summary", "Status"})

	var index = 0

	for i := range set {
		index = index + 1
		if _, ok := l1.deployment[i]; !ok {
			string := fmt.Sprintf("%v has: %v\n%v has: %v", c1, l1.deployment[i].Name, c2, l2.deployment[i].Name)
			t.AppendRows([]table.Row{
				{index, i, string, "X"},
			})
		} else if _, ok := l2.deployment[i]; !ok {
			string := fmt.Sprintf("%v has: %v\n%v has: %v", c1, l1.deployment[i].Name, c2, l2.deployment[i].Name)
			t.AppendRows([]table.Row{
				{index, i, string, "X"},
			})
		} else if l1.deployment[i].Name != l2.deployment[i].Name {
			string := fmt.Sprintf("%v has: %v\n%v has: %v", c1, l1.deployment[i].Name, c2, l2.deployment[i].Name)
			t.AppendRows([]table.Row{
				{index, i, string, "X"},
			})
		} else {
			t.AppendRows([]table.Row{
				{index, i, "", "X"},
			})
		}
	}
	t.Render()
}

func main() {
	compareComponents("", "", "deployment")
}
