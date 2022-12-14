package chk_components

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

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
	clientSet     *kubernetes.Clientset
	kubeconfig    *string
	Clustersarg   []string
	Namespacesarg []string
	Resourcesarg  string
)

func init() {
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	clusters := flag.String("c", "", "Provide cluster name you want to check. ( eg. -c=test1,test2 )")
	resources := flag.String("r", "", "Provide resources type you want to check. ( eg. -r=pod )")
	namespaces := flag.String("n", "default", "Provide namespaces you want to check. ( eg. -r=caas-system,kube-system )")

	flag.Parse()

	Clustersarg = strings.Split(*clusters, ",")
	Namespacesarg = strings.Split(*namespaces, ",")
	Resourcesarg = *resources

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

type Label struct {
	LabelName string
}

type LabelList struct {
	LabelList []Label
}

type ClusterLabel struct {
	Cluster map[string]LabelList
}

func splitStrings(name string) string {
	var splitLines []string
	strLength := len(name) - 1
	gap := 30
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
	var currentcontext string
	for _, c := range clusters {
		currentcontext = GetConfigFromConfig(c, *kubeconfig)
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
		summary = append(summary, splitStrings(i))
		var imageArray [][]string
		for _, k := range clusterkeys {
			imageLists := make([]string, 0, len(l.Clusters[k].Resource[i]))
			for _, k := range l.Clusters[k].Resource[i] {
				imageLists = append(imageLists, splitStrings(k.Name))
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
