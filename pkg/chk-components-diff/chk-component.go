package chk_components

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	clientSet  *kubernetes.Clientset
	kubeconfig *string
	// Clusters     *string
	Clustersarg []string
	// Resources    *string
	Resourcesarg string
)

func init() {
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	Clusters := flag.String("c", "", "Provide cluster name you want to check the diff. ( eg. -c=test1,test2 )")
	Resources := flag.String("r", "", "Provide resources type you want to check the diff. ( eg. -r=pod )")

	flag.Parse()

	Clustersarg = strings.Split(*Clusters, ",")
	Resourcesarg = *Resources

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

func GetNodes() {

	resource, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
		// LabelSelector: "cluster.aps.cpd.rakuten.com/rackInfo",
	})
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	tr := table.Row{"Nodes", "Label"}
	t.AppendHeader(tr)
	if err != nil {
		panic(err.Error())
	}
	for _, d := range resource.Items {
		bindingTaint := "Missing"
		for b := range d.Labels {
			if b == "cluster.aps.cpd.rakuten.com/rackInfo" {
				bindingTaint = d.Labels["cluster.aps.cpd.rakuten.com/rackInfo"]
				break
			}
		}
		// for _, b := range d.Spec.Taints {
		// 	if b.Key == "network.cpd.rakuten.com/dlb-binding" {
		// 		bindingTaint = b.Value
		// 	}
		// }
		if bindingTaint == "Missing" {
			fmt.Printf("%v, %v\n", d.Name, bindingTaint)
		}
		// t.AppendRow([]interface{}{d.Name, bindingTaint})
		// t.AppendSeparator()
	}
	// t.Render()
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
		// for _, c := range clusters {
		// 	currentcontext = GetConfigFromConfig(c, *kubeconfig)
		// 	switchContext(currentcontext)
		// 	list := GetPod()
		// 	for i := range list.Resource {
		// 		if !set[i] {
		// 			set[i] = true
		// 		}
		// 	}
		// 	l.Clusters[c] = list
		// }
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	tr := table.Row{"Resource"}
	keys := make([]string, 0, len(l.Clusters))
	for _, c := range clusters {
		tr = append(tr, c)
		keys = append(keys, c)
	}
	tr = append(tr, "Status")
	t.AppendHeader(tr)

	for i := range set {
		summary := []string{}

		var flag bool
		count := 0
		summary = append(summary, splitStrings(i))
		var imageArray [][]string
		for _, k := range keys {
			count++
			keys := make([]string, 0, len(l.Clusters[k].Resource[i]))
			for _, k := range l.Clusters[k].Resource[i] {
				keys = append(keys, splitStrings(k.Name))
			}

			sort.Strings(keys)
			imageArray = append(imageArray, keys)
			summary = append(summary, fmt.Sprintf("%v", strings.Join(keys, "\n")))
			t.AppendSeparator()
			if _, ok := l.Clusters[k].Resource[i]; !ok {
				flag = true
			} else if !reflect.DeepEqual(imageArray[0], keys) {
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
