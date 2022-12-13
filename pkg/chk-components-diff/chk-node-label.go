package chk_components

import (
	"context"
	"fmt"
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

func GetNodes(clusters ...string) ClusterLabel {
	var clusterLabel ClusterLabel
	clusterLabel.Cluster = map[string]LabelList{}
	var currentcontext string
	for _, c := range clusters {
		currentcontext = GetConfigFromConfig(c, *kubeconfig)
		switchContext(currentcontext)
		resource, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
			LabelSelector: "node-role.kubernetes.io/master=",
		})
		if err != nil {
			panic(err.Error())
		}
		var llist LabelList
		for _, d := range resource.Items {
			for b := range d.Labels {
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

func CompareLabels(clusters ...string) {
	labellist := GetNodes(clusters...)
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

	for _, c := range clusters {
		tr = append(tr, c)
	}
	tr = append(tr, "Status")
	t.AppendHeader(tr)

	for r := range set {
		fmt.Printf("%v\n", r)
		flag := true
		summary := []string{}
		for _, l := range labellist.Cluster {
			for _, list := range l.LabelList {
				if list.LabelName == r {
					summary = append(summary, list.LabelName)
					flag = false
				}
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
