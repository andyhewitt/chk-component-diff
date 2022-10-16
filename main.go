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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
	container map[string]ContainerInfo
}

type ContainerInfo struct {
	Name      string
	Namespace string
}

func getDeployment() ContainerList {
	cl := ContainerList{
		container: map[string]ContainerInfo{},
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
				separateImageRegex := regexp.MustCompile(".+/(.+):(.+)")
				rs := separateImageRegex.FindStringSubmatch(imageName)
				fmt.Printf("%v, %v, %v\n", rs[0], rs[1], rs[2])
				cl.container[containerName] = ContainerInfo{
					Name:      m.ReplaceAllString(imageName, ""),
					Namespace: ns,
				}
			}
		}
	}
	return cl
}

func getPod() ContainerList {
	cl := ContainerList{
		container: map[string]ContainerInfo{},
	}

	namespaces := []string{
		"kube-system",
	}
	for _, ns := range namespaces {
		resource, err := clientSet.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, d := range resource.Items {
			for _, c := range d.Spec.Containers {
				imageName := c.Image
				containerName := c.Name
				m := regexp.MustCompile("^registry.+net/")
				cl.container[containerName] = ContainerInfo{
					Name:      m.ReplaceAllString(imageName, ""),
					Namespace: ns,
				}
			}
		}
	}
	return cl
}

func compareComponents(n string, clusters ...string) {
	var l []ContainerList

	var currentcontext string
	set := make(map[string]bool)

	switch n {
	case "deployment":
		fmt.Printf("Pass")
		// currentcontext = getConfigFromConfig(c1, *kubeconfig)
		// switchContext(currentcontext)
		// l1 = getDeployment()
		// currentcontext = getConfigFromConfig(c2, *kubeconfig)
		// switchContext(currentcontext)
		// l2 = getDeployment()
	case "pod":
		for _, c := range clusters {
			currentcontext = getConfigFromConfig(c, *kubeconfig)
			switchContext(currentcontext)
			list := getPod()
			fmt.Print(list)
			for i := range list.container {
				if !set[i] {
					set[i] = true
				}
			}
			l = append(l, list)
		}
	}

	// for i := range l2.container {
	// 	if !set[i] {
	// 		set[i] = true
	// 	}
	// }

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Resource", "Summary", "Status"})

	var index = 0

	for i := range set {
		index = index + 1
		var stringList []string
		for _, c := range l {
			if _, ok := c.container[i]; !ok {
				string := fmt.Sprintf("%v has: %v\n", c, c.container[i].Name)
				stringList = append(stringList, string)
			}
		}
		t.AppendRows([]table.Row{
			{index, i, strings.Join(stringList, ""), "🥹"},
		})
		// if _, ok := l1.container[i]; !ok {
		// 	string := fmt.Sprintf("%v has: %v\n%v has: %v", c1, l1.container[i].Name, c2, l2.container[i].Name)
		// 	t.AppendRows([]table.Row{
		// 		{index, i, string, "🥹"},
		// 	})
		// } else if _, ok := l2.container[i]; !ok {
		// 	string := fmt.Sprintf("%v has: %v\n%v has: %v", c1, l1.container[i].Name, c2, l2.container[i].Name)
		// 	t.AppendRows([]table.Row{
		// 		{index, i, string, "🥹"},
		// 	})
		// } else if l1.container[i].Name != l2.container[i].Name {
		// 	string := fmt.Sprintf("%v has: %v\n%v has: %v", c1, l1.container[i].Name, c2, l2.container[i].Name)
		// 	t.AppendRows([]table.Row{
		// 		{index, i, string, "🥹"},
		// 	})
		// } else {
		//     string := fmt.Sprintf("%v has: %v\n%v has: %v", c1, l1.container[i].Name, c2, l2.container[i].Name)
		// 	t.AppendRows([]table.Row{
		// 		{index, i, string, "😊"},
		// 	})
		// }
	}
	t.Render()
}

func main() {
	compareComponents("pod", "test1", "test2")
}
