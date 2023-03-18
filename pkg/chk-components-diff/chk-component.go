package chk_components

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"k8s.io/client-go/kubernetes"

	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func switchContext(context string) {
	// use the current context in kubeconfig
	config, err := BuildConfigFromFlags(context, *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
}

func SplitStrings(name string, gap int) string {
	var splitLines []string
	strLength := len(name) - 1
	if strLength <= gap {
		return name
	}
	start := 0
	end := start + gap
	for end < strLength {
		if start+gap < strLength {
			end = start + gap
		} else {
			end = strLength + 1
		}
		l := name[start:end]
		start = end
		splitLines = append(splitLines, l)
	}
	return fmt.Sprintf("%v", strings.Join(splitLines, "\n"))
}

func processResourceList(resource string, set map[string]map[string]bool, l *ClusterContainers, clusters []string) (map[string]map[string]bool, ClusterContainers) {

	set[resource] = make(map[string]bool)

	for _, cluster := range clusters {
		currentcontext, err := GetConfigFromConfig(cluster, *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		switchContext(currentcontext)
		var list ResourceList
		switch resource {
		case "deployment", "deploy":
			list = GetDeployment()
		case "daemonset", "ds":
			list = GetDaemonSets()
		case "statefulset", "sts":
			list = GetStatefulSets()
		case "pod", "po":
			list = GetPod()
		}

		for i := range list.ResourceName {
			if !set[resource][i] {
				set[resource][i] = true
			}
		}

		resourceType := ResourceType{
			Resource: map[string]ResourceList{},
		}
		resourceType.Resource[resource] = list
		l.Clusters[cluster] = resourceType
	}
	return set, *l
}

func CompareComponents(resourceType []string, clusters []string) {

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	tr := table.Row{"Resource"}
	tr = append(tr, "ResourceType")
	tr = append(tr, "Namespace")
	clusterkeys := make([]string, 0, len(clusters))
	for _, c := range clusters {
		tr = append(tr, c)
		clusterkeys = append(clusterkeys, c)
	}
	tr = append(tr, "Status")
	t.AppendHeader(tr)
	for _, resource := range resourceType {
		l := ClusterContainers{
			Clusters: map[string]ResourceType{},
		}

		var resourceSet = make(map[string]map[string]bool)

		switch resource {
		case "deployment", "deploy":
			resourceSet, l = processResourceList(resource, resourceSet, &l, clusters)
		case "daemonset", "ds":
			resourceSet, l = processResourceList(resource, resourceSet, &l, clusters)
		case "pod", "po":
			resourceSet, l = processResourceList(resource, resourceSet, &l, clusters)
		case "statefulset", "sts":
			resourceSet, l = processResourceList(resource, resourceSet, &l, clusters)
		}

		container := resourceSet[resource]
		for i := range container {
			summary := []string{}

			var flag bool
			flag = false
			summary = append(summary, SplitStrings(i, 30))
			summary = append(summary, SplitStrings(resource, 30))
			var imageArray [][]string
			for _, k := range clusterkeys {
				summary = append(summary, SplitStrings(l.Clusters[k].Resource[resource].ResourceName[i].Namespace, 30))
				currentResource := l.Clusters[k].Resource[resource].ResourceName[i]
				imageLists := make([]string, 0, len(currentResource.Container))
				for _, c := range currentResource.Container {
					imageLists = append(imageLists, SplitStrings(c.LongImageName, 30))
				}

				sort.Strings(imageLists)
				imageArray = append(imageArray, imageLists)
				summary = append(summary, fmt.Sprintf("%v", strings.Join(imageLists, "\n")))
				// fmt.Printf("%+v\n", summary)
				t.AppendSeparator()
				if _, ok := l.Clusters[k].Resource[resource].ResourceName[i]; !ok {
					flag = true
				} else if !reflect.DeepEqual(imageArray[0], imageLists) {
					flag = true
				}
			}

			if flag {
				summary = append(summary, "💀")
			} else {
				summary = append(summary, "😄")
			}
			rest := table.Row{}
			for _, m := range summary {
				rest = append(rest, m)
			}
			t.AppendRows([]table.Row{
				rest,
			})
		}
	}
	t.SetAutoIndex(true)
	// t.SortBy([]table.SortBy{
	// 	{Name: "Resource", Mode: table.Asc},
	// })
	t.Render()
}
