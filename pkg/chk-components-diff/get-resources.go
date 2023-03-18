package chk_components

import (
	"context"
	"regexp"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


type ContainerInfo struct {
	LongImageName      string
	Namespace string
	Registry  string
	Image     string
	Version   string
}

// For multiple container Pod/Deployment etc
type ContainerName struct {
	ContainerName map[string]ContainerInfo
}

type ResourceList struct {
	ResourceName map[string]ContainerName
}

type ResourceType struct {
	Resource map[string]ResourceList
}

type ClusterContainers struct {
	Clusters map[string]ResourceType
}


func processResource(rl *ResourceList, resourceName string, ns string, containers []v1.Container) {
	for _, c := range containers {
		cName := c.Name
		imageName := c.Image
		m := regexp.MustCompile("^registry.+net/")
		separateImageRegex := regexp.MustCompile("(.+/)(.+):(.+)")
		rs := separateImageRegex.FindStringSubmatch(imageName)
		if len(rs) < 3 {
			rl.ResourceName[resourceName].ContainerName[cName] = ContainerInfo{
				LongImageName:      imageName,
				Namespace: ns,
				Registry:  "",
				Image:     imageName,
				Version:   "",
			}
		} else {
			// CaaS format
			rl.ResourceName[resourceName].ContainerName[cName] = ContainerInfo{
				LongImageName:      m.ReplaceAllString(imageName, ""),
				Namespace: ns,
				Registry:  rs[1],
				Image:     rs[2],
				Version:   rs[3],
			}
		}
	}
}

func GetDeployment() ResourceList {
	// create an empty variable of the ResourceList struct type
	rl := ResourceList{}

	// initialize the ResourceName field with a map
	rl.ResourceName = make(map[string]ContainerName)

	namespaces := Namespacesarg
	for _, ns := range namespaces {

		resource, err := clientSet.AppsV1().Deployments(ns).List(context.TODO(), metav1.ListOptions{})
		// resourceName := resource.ma
		if err != nil {
			panic(err.Error())
		}
		for _, d := range resource.Items {
			resourceName := d.Name
			// add a new resource to the map
			cn := ContainerName{}
			cn.ContainerName = make(map[string]ContainerInfo)
			rl.ResourceName[d.Name] = cn
			processResource(&rl, resourceName, ns, d.Spec.Template.Spec.Containers)
		}
	}
	
	return rl
}

func GetDaemonSets() ResourceList {
	// create an empty variable of the ResourceList struct type
	rl := ResourceList{}

	// initialize the ResourceName field with a map
	rl.ResourceName = make(map[string]ContainerName)

	namespaces := Namespacesarg
	for _, ns := range namespaces {
		resource, err := clientSet.AppsV1().DaemonSets(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, d := range resource.Items {
			resourceName := d.Name
			// add a new resource to the map
			cn := ContainerName{}
			cn.ContainerName = make(map[string]ContainerInfo)
			rl.ResourceName[d.Name] = cn
			processResource(&rl, resourceName, ns, d.Spec.Template.Spec.Containers)
		}
	}
	return rl
}

func GetStatefulSets() ResourceList {
	// create an empty variable of the ResourceList struct type
	rl := ResourceList{}

	// initialize the ResourceName field with a map
	rl.ResourceName = make(map[string]ContainerName)

	namespaces := Namespacesarg
	for _, ns := range namespaces {
		resource, err := clientSet.AppsV1().StatefulSets(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, d := range resource.Items {
			resourceName := d.Name
			// add a new resource to the map
			cn := ContainerName{}
			cn.ContainerName = make(map[string]ContainerInfo)
			rl.ResourceName[d.Name] = cn
			processResource(&rl, resourceName, ns, d.Spec.Template.Spec.Containers)
		}
	}
	return rl
}

func GetPod() ResourceList {
	// create an empty variable of the ResourceList struct type
	rl := ResourceList{}

	// initialize the ResourceName field with a map
	rl.ResourceName = make(map[string]ContainerName)

	namespaces := Namespacesarg
	for _, ns := range namespaces {
		resource, err := clientSet.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, d := range resource.Items {
			resourceName := d.Name
			// add a new resource to the map
			cn := ContainerName{}
			cn.ContainerName = make(map[string]ContainerInfo)
			rl.ResourceName[d.Name] = cn
			processResource(&rl, resourceName, ns, d.Spec.Containers)
		}
	}

	return rl
}
