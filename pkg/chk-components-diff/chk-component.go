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

func SplitStrings(name string) string {
	gap := TableLengtharg
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

func processResourceList(resource string, set map[string]map[string]bool, cc *ClusterContainers, clusters []string) (map[string]map[string]bool, ClusterContainers) {

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
		cc.Clusters[cluster] = resourceType
	}
	return set, *cc
}

func PrepareCompareRawData(resource string, clusters []string) (map[string]map[string]bool, ClusterContainers) {
	cc := ClusterContainers{
		Clusters: map[string]ResourceType{},
	}

	resourceSet := make(map[string]map[string]bool)

	switch resource {
	case "deployment", "deploy", "daemonset", "ds", "pod", "po", "statefulset", "sts":
		resourceSet, cc = processResourceList(resource, resourceSet, &cc, clusters)
	}

	return resourceSet, cc
}

func CompareComponents(resourceType []string, clusters []string) {

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	tr := table.Row{"Resource", "ResourceType", "Namespace"}
	for _, c := range clusters {
		tr = append(tr, c)
	}
	tr = append(tr, "Status")
	t.AppendHeader(tr)
	for _, resource := range resourceType {
		resourceSet, cc := PrepareCompareRawData(resource, clusters)

		containers := resourceSet[resource]
		for container := range containers {
			summary := []string{}
			summary = append(summary, SplitStrings(container))
			summary = append(summary, SplitStrings(resource))

			var imageArray [][]string
			var nameSpace string
			for _, k := range clusters {

				currentResource := cc.Clusters[k].Resource[resource]
				imageLists := make([]string, 0, len(currentResource.ResourceName[container].Container))

				if currentResource, ok := currentResource.ResourceName[container]; ok {
					nameSpace = currentResource.Namespace
				}

				for _, c := range currentResource.ResourceName[container].Container {
					imageLists = append(imageLists, SplitStrings(c.LongImageName))
				}

				sort.Strings(imageLists)
				imageArray = append(imageArray, imageLists)

				t.AppendSeparator()
			}

			summary = append(summary, nameSpace)
			if !AllEqual(imageArray) {
				summary = append(summary, imagesToSummary(imageArray)...)
				summary = append(summary, "💀")
			} else {
				summary = append(summary, imagesToSummary(imageArray)...)
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
	t.SortBy([]table.SortBy{
		{Name: "ResourceType", Mode: table.Asc},
		{Name: "Namespace", Mode: table.Asc},
		{Name: "Resource", Mode: table.Asc},
	})
	t.Render()
}

func imagesToSummary(images [][]string) []string {
	var summary []string
	for _, imageList := range images {
		if len(imageList) == 0 {
			summary = append(summary, "-")
		} else {
			summary = append(summary, strings.Join(imageList, "\n"))
		}
	}
	// fmt.Printf("%+v\n", summary)
	return summary
	// return strings.Join(summary, "\n")
}

func AllEqual(arr [][]string) bool {
	for i := 1; i < len(arr); i++ {
		if !reflect.DeepEqual(arr[0], arr[i]) {
			return false
		}
	}
	return true
}
