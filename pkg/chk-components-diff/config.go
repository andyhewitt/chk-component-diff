package chk_components

import (
	"flag"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	clientSet      *kubernetes.Clientset
	kubeconfig     *string
	Clustersarg    []string
	Namespacesarg  []string
	Resourcesarg   []string
	Labelarg       string
	TableLengtharg int
	CaaSCheck      bool
)

func init() {
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	clusters := flag.String("c", "", "Provide cluster name you want to check. ( eg. -c=test1,test2 )")
	resources := flag.String("r", "", "Provide resources type you want to check. ( eg. -r=deploy,sts )")
	namespaces := flag.String("n", "default", "Provide namespaces you want to check. ( eg. -r=caas-system,kube-system )")
	label := flag.String("l", "", "Provide a label you want to check. ( eg. -l=cluster.aps.cpd.rakuten.com/noderole=master )")
	flag.IntVar(&TableLengtharg, "table-length", 30, "The maximum word length to show on a singel cell. ( eg. -table-length=30 )")

	flag.Parse()

	Clustersarg = strings.Split(*clusters, ",")
	Namespacesarg = strings.Split(*namespaces, ",")
	Resourcesarg = strings.Split(*resources, ",")
	Labelarg = *label
	// TableLengtharg = tableLength

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

func BuildConfigFromFlags(context, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

func GetConfigFromConfig(context, kubeconfigPath string) (string, error) {
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{})
	clientConfig, err := config.RawConfig()
	if err != nil {
		panic(err.Error())
	}
	clusterRegex := fmt.Sprintf("^%s.*", context)
	for n := range clientConfig.Contexts {
		re := regexp.MustCompile(clusterRegex)
		result := re.MatchString(n)
		if result {
			return re.FindString(n), nil
		}
	}
	return "", fmt.Errorf("cannot find context %s", context)
}
