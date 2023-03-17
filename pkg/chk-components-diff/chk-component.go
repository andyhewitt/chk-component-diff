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

func processResourceList(resource string, set map[string]bool, l *ClusterContainers, clusters ...string) (map[string]bool, ClusterContainers) {
	for _, c := range clusters {
		currentcontext, err := GetConfigFromConfig(c, *kubeconfig)
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
		// list := GetDaemonSets()
		for i := range list.Resource {
			if !set[i] {
				set[i] = true
			}
		}
		l.Clusters[c] = list
	}
	return set, *l
}

func CompareComponents(n string, clusters ...string) {
	l := ClusterContainers{
		Clusters: map[string]ResourceList{},
	}

	var set = make(map[string]bool)

	switch n {
	case "deployment", "deploy":
		set, l = processResourceList(n, set, &l, clusters...)
	case "daemonset", "ds":
		set, l = processResourceList(n, set, &l, clusters...)
	case "pod", "po":
		set, l = processResourceList(n, set, &l, clusters...)
	case "statefulset", "sts":
		set, l = processResourceList(n, set, &l, clusters...)
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	tr := table.Row{"Resource"}
	clusterkeys := make([]string, 0, len(l.Clusters))
	for _, c := range clusters {
		tr = append(tr, c)
		clusterkeys = append(clusterkeys, c)
	}
	tr = append(tr, "Status")
	t.AppendHeader(tr)

	for i := range set {
		summary := []string{}

		var flag bool
		summary = append(summary, SplitStrings(i, 30))
		var imageArray [][]string
		for _, k := range clusterkeys {
			imageLists := make([]string, 0, len(l.Clusters[k].Resource[i]))
			for _, k := range l.Clusters[k].Resource[i] {
				imageLists = append(imageLists, SplitStrings(k.Name, 30))
			}

			sort.Strings(imageLists)
			imageArray = append(imageArray, imageLists)
			summary = append(summary, fmt.Sprintf("%v", strings.Join(imageLists, "\n")))
			t.AppendSeparator()
			if _, ok := l.Clusters[k].Resource[i]; !ok {
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
	t.SetAutoIndex(true)
	t.SortBy([]table.SortBy{
		{Name: "Resource", Mode: table.Asc},
	})
	t.Render()
}
