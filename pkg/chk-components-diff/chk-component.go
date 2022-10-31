package chk_components

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
	clientSet   *kubernetes.Clientset
	kubeconfig  *string
	Clusters    *string
	Clustersarg []string
)

func init() {
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	Clusters = flag.String("c", "", "Provide cluster name you want to check the diff. ( eg. -c=test1,test2 )")

	flag.Parse()

	Clustersarg = strings.Split(*Clusters, ",")

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
		LabelSelector: "minikube.k8s.io/name",
	})
	if err != nil {
		panic(err.Error())
	}
	for _, d := range resource.Items {
		fmt.Printf("%v\n", d)
	}
}

func CompareComponents(n string, clusters ...string) {
	l := ClusterContainers{
		Clusters: map[string]ResourceList{},
	}

	var currentcontext string
	set := make(map[string]bool)

	switch n {
	case "deployment":
		for _, c := range clusters {
			currentcontext = GetConfigFromConfig(c, *kubeconfig)
			switchContext(currentcontext)
			list := GetDeployment()
			for i := range list.Resource {
				if !set[i] {
					set[i] = true
				}
			}
			l.Clusters[c] = list
		}
	case "pod":
		for _, c := range clusters {
			currentcontext = GetConfigFromConfig(c, *kubeconfig)
			switchContext(currentcontext)
			list := GetPod()
			for i := range list.Resource {
				if !set[i] {
					set[i] = true
				}
			}
			l.Clusters[c] = list
		}
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	tr := table.Row{"#", "Resource"}
	keys := make([]string, 0, len(l.Clusters))
	for _, c := range clusters {
		tr = append(tr, c)
		keys = append(keys, c)
	}
	tr = append(tr, "Status")
	t.AppendHeader(tr)

	var index = 0

	for i := range set {
		summary := []string{}
		index = index + 1

		var flag bool
		count := 0
		summary = append(summary, strconv.Itoa(index), i)
		for _, k := range keys {
			count++
			summary = append(summary, fmt.Sprintf("%v\n%v", l.Clusters[k].Resource[i].Image, l.Clusters[k].Resource[i].Version))
			t.AppendSeparator()
			if _, ok := l.Clusters[k].Resource[i]; !ok {
				flag = true
			} else if l.Clusters[keys[0]].Resource[i].Name != l.Clusters[k].Resource[i].Name {
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
		// fmt.Printf("%v\n", summary)
	}
	t.Render()
}
