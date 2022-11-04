package chk_components

import (
	"context"
	"regexp"

	v1 "k8s.io/api/core/v1"
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

func GetDeployment() ResourceList {
	cl := ResourceList{
		Resource: map[string]map[string]ContainerInfo{},
	}

	namespaces := []string{
		// "kube-system",
		"caas-system",
	}
	for _, ns := range namespaces {
		resource, err := clientSet.AppsV1().Deployments(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, d := range resource.Items {
			resourceName := d.Name
			cl.Resource[d.Name] = map[string]ContainerInfo{}
			processResource(cl, resourceName, ns, d.Spec.Template.Spec.Containers)
		}
	}
	return cl
}

func processResource(cl ResourceList, resourceName string, ns string, containers []v1.Container) {
	for _, c := range containers {
		containerName := c.Name
		imageName := c.Image
		m := regexp.MustCompile("^registry.+net/")
		separateImageRegex := regexp.MustCompile("(.+/)(.+):(.+)")
		rs := separateImageRegex.FindStringSubmatch(imageName)
		if len(rs) < 3 {
			cl.Resource[resourceName][containerName] = ContainerInfo{
				Name:      imageName,
				Namespace: ns,
				Registry:  "",
				Image:     imageName,
				Version:   "",
			}
		} else {
			cl.Resource[resourceName][containerName] = ContainerInfo{
				Name:      m.ReplaceAllString(imageName, ""),
				Namespace: ns,
				Registry:  rs[1],
				Image:     rs[2],
				Version:   rs[3],
			}
		}
	}
}

func GetDaemonSets() ResourceList {
	cl := ResourceList{
		Resource: map[string]map[string]ContainerInfo{},
	}

	namespaces := []string{
		"kube-system",
	}
	for _, ns := range namespaces {
		resource, err := clientSet.AppsV1().DaemonSets(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, d := range resource.Items {
			// fmt.Printf("%v\n", d.Name)
			resourceName := d.Name
			cl.Resource[resourceName] = map[string]ContainerInfo{}
			processResource(cl, resourceName, ns, d.Spec.Template.Spec.Containers)
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
			resourceName := d.Name
			cl.Resource[d.Name] = map[string]ContainerInfo{}
			processResource(cl, resourceName, ns, d.Spec.Containers)
		}
	}
	return cl
}
