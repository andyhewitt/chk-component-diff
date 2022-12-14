package release_test

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
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

func TestSum(t *testing.T) {
	resource, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
		// LabelSelector: "cluster.aps.cpd.rakuten.com/rackInfo",
	})

	if err != nil {
		panic(err.Error())
	}
	targetLabel := "cluster.aps.cpd.rakuten.com/rackInfo"
	for _, d := range resource.Items {
		for b := range d.Labels {
			fmt.Print(b)
			if b == targetLabel {
				break
			}
		}
	}
	t.Error(true, true)
}
