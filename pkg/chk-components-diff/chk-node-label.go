package chk_components

import (
	"context"
	"math/rand"
	"os"

	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/jedib0t/go-pretty/v6/table"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

type Label struct {
	LabelName string
}

type LabelList struct {
	LabelList []Label
}

type ClusterLabel struct {
	Cluster map[string]LabelList
}

func GetNodes(label string, clusters []string) ClusterLabel {
	var clusterLabel ClusterLabel
	clusterLabel.Cluster = map[string]LabelList{}
	for _, c := range clusters {
		currentcontext, err := GetConfigFromConfig(c, *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		switchContext(currentcontext)
		resource, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
			LabelSelector: label,
		})
		if err != nil {
			panic(err.Error())
		}
		var llist LabelList
		if len(resource.Items) != 0 {
			randInx := rand.Intn(len(resource.Items))
			for b := range resource.Items[randInx].Labels {
				var l = Label{
					LabelName: b,
				}
				llist.LabelList = append(llist.LabelList, l)
			}
		}
		clusterLabel.Cluster[c] = llist
	}
	return clusterLabel
}

func CompareLabels(label string, clusters []string) {
	labellist := GetNodes(label, clusters)
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	tr := table.Row{}
	set := make(map[string]bool)
	for _, l := range labellist.Cluster {
		for _, list := range l.LabelList {
			if !set[list.LabelName] {
				set[list.LabelName] = true
			}
		}
	}
	tr = append(tr, "Item")
	for _, c := range clusters {
		tr = append(tr, c)
	}
	t.AppendHeader(tr)

	for item := range set {
		summary := []string{}
		summary = append(summary, SplitStrings(item, 30))
		for _, c := range clusters {
			flag := true
			for _, l := range labellist.Cluster[c].LabelList {
				if l.LabelName == item {
					flag = false
					break
				}
			}
			if flag {
				summary = append(summary, "💀")
			} else {
				summary = append(summary, "😄")
			}
		}
		rest := table.Row{}
		for _, m := range summary {
			rest = append(rest, m)
		}
		t.AppendRows([]table.Row{
			rest,
		})
		t.AppendSeparator()
	}
	t.SetAutoIndex(true)
	t.SortBy([]table.SortBy{
		{Name: "Resource", Mode: table.Asc},
	})
	t.Render()

}
