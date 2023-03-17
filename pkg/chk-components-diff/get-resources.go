package chk_components

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


type ContainerInfo struct {
	Name      string
	Namespace string
	Registry  string
	Image     string
	Version   string
}

type ResourceType struct {
	Resource map[string]ResourceList
}

type ResourceList struct {
	ResourceName map[string]map[string]ContainerInfo
}

type ClusterContainers struct {
	Clusters map[string]ResourceType
}


func processResource(cl *ResourceList, name string, ns string, containers []v1.Container) {
	for _, c := range containers {
		containerName := c.Name
		imageName := c.Image
		m := regexp.MustCompile("^registry.+net/")
		separateImageRegex := regexp.MustCompile("(.+/)(.+):(.+)")
		rs := separateImageRegex.FindStringSubmatch(imageName)
		if len(rs) < 3 {
			cl.ResourceName[name][containerName] = ContainerInfo{
				Name:      imageName,
				Namespace: ns,
				Registry:  "",
				Image:     imageName,
				Version:   "",
			}
		} else {
			cl.ResourceName[name][containerName] = ContainerInfo{
				Name:      m.ReplaceAllString(imageName, ""),
				Namespace: ns,
				Registry:  rs[1],
				Image:     rs[2],
				Version:   rs[3],
			}
		}
	}
}

func GetDeployment() ResourceList {
	cl := ResourceList{
		ResourceName: map[string]map[string]ContainerInfo{},
	}

	namespaces := Namespacesarg
	for _, ns := range namespaces {
		resource, err := clientSet.AppsV1().Deployments(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, d := range resource.Items {
			resourceName := d.Name
			cl.ResourceName[d.Name] = map[string]ContainerInfo{}
			processResource(&cl, resourceName, ns, d.Spec.Template.Spec.Containers)
		}
	}
	
    b, err := json.MarshalIndent(cl, "", "    ")
    if err != nil {
        fmt.Println("Error:", err)
    }
    fmt.Println(string(b))
	return cl
}

func GetDaemonSets() ResourceList {
	cl := ResourceList{
		ResourceName: map[string]map[string]ContainerInfo{},
	}

	namespaces := Namespacesarg
	for _, ns := range namespaces {
		resource, err := clientSet.AppsV1().DaemonSets(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, d := range resource.Items {
			// fmt.Printf("%v\n", d.Name)
			name := d.Name
			cl.ResourceName[name] = map[string]ContainerInfo{}
			processResource(&cl, name, ns, d.Spec.Template.Spec.Containers)
		}
	}
	return cl
}

func GetStatefulSets() ResourceList {
	cl := ResourceList{
		ResourceName: map[string]map[string]ContainerInfo{},
	}

	namespaces := Namespacesarg
	for _, ns := range namespaces {
		resource, err := clientSet.AppsV1().StatefulSets(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, d := range resource.Items {
			// fmt.Printf("%v\n", d.Name)
			name := d.Name
			cl.ResourceName[name] = map[string]ContainerInfo{}
			processResource(&cl, name, ns, d.Spec.Template.Spec.Containers)
		}
	}
	return cl
}

func GetPod() ResourceList {
	// cl := ResourceList{}
	cl := ResourceList{
		ResourceName: map[string]map[string]ContainerInfo{},
	}

	namespaces := Namespacesarg
	for _, ns := range namespaces {
		resource, err := clientSet.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, d := range resource.Items {
			name := d.Name
			cl.ResourceName[d.Name] = map[string]ContainerInfo{}
			processResource(&cl, name, ns, d.Spec.Containers)
		}
	}

	b, err := json.MarshalIndent(cl, "", "    ")
    if err != nil {
        fmt.Println("Error:", err)
    }
    fmt.Println(string(b))
	return cl
}
