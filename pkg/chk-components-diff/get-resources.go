package chk_components

import (
	"context"
	"fmt"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


type ResourceList struct {
	Resource map[string]map[string]ContainerInfo
}

type ClusterContainers struct {
	Clusters map[string]ResourceList
}

type ContainerInfo struct {
	Name      string
	Namespace string
	Registry  string
	Image     string
	Version   string
}

// coredns: {
//     coredns: {
//         name: xxx
//     },
//     proxy: {
//         name: xxx
//     }
// }

func GetDeployment() ResourceList {
	cl := ResourceList{
		Resource: map[string]map[string]ContainerInfo{},
	}

	namespaces := []string{
		"kube-system",
	}
	for _, ns := range namespaces {
		resource, err := clientSet.AppsV1().Deployments(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, d := range resource.Items {
            fmt.Printf("%v\n", d.Name)
            cl.Resource[d.Name] = map[string]ContainerInfo{}
			for _, c := range d.Spec.Template.Spec.Containers {
				containerName := c.Name
				imageName := c.Image
				m := regexp.MustCompile("^registry.+net/")
				separateImageRegex := regexp.MustCompile("(.+/)(.+):(.+)")
				rs := separateImageRegex.FindStringSubmatch(imageName)
				if len(rs) < 3 {
					cl.Resource[d.Name][containerName] = ContainerInfo{
						Name:      imageName,
						Namespace: ns,
						Registry:  "",
						Image:     imageName,
						Version:   "",
					}
				} else {
					cl.Resource[d.Name][containerName] = ContainerInfo{
						Name:      m.ReplaceAllString(imageName, ""),
						Namespace: ns,
						Registry:  rs[1],
						Image:     rs[2],
						Version:   rs[3],
					}
				}
			}
		}
	}
	return cl
}

func GetPod() ResourceList {
	cl := ResourceList{
		Resource: map[string]map[string]ContainerInfo{},
	}

	namespaces := []string{
		// "kube-system",
        "default",
	}
	for _, ns := range namespaces {
		resource, err := clientSet.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, d := range resource.Items {
            cl.Resource[d.Name] = map[string]ContainerInfo{}
			for _, c := range d.Spec.Containers {
				imageName := c.Image
				containerName := c.Name
				m := regexp.MustCompile("^registry.+net/")
				separateImageRegex := regexp.MustCompile("(.+/)(.+):(.+)")
				rs := separateImageRegex.FindStringSubmatch(imageName)
				if len(rs) < 3 {
					cl.Resource[d.Name][containerName] = ContainerInfo{
						Name:      m.ReplaceAllString(imageName, ""),
						Namespace: ns,
						Registry:  "",
						Image:     imageName,
						Version:   "",
					}
				} else {
					cl.Resource[d.Name][containerName] = ContainerInfo{
						Name:      m.ReplaceAllString(imageName, ""),
						Namespace: ns,
						Registry:  rs[1],
						Image:     rs[2],
						Version:   rs[3],
					}
				}
			}
		}
	}
	return cl
}