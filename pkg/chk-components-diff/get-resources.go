package chk_components

import (
	"context"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetDeployment() ContainerList {
	cl := ContainerList{
		Container: map[string]ContainerInfo{},
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
			for _, c := range d.Spec.Template.Spec.Containers {
				imageName := c.Image
				containerName := c.Name
				m := regexp.MustCompile("^registry.+net/")
				separateImageRegex := regexp.MustCompile("(.+/)(.+):(.+)")
				rs := separateImageRegex.FindStringSubmatch(imageName)
				if len(rs) < 3 {
					cl.Container[containerName] = ContainerInfo{
						Name:      imageName,
						Namespace: ns,
						Registry:  "",
						Image:     imageName,
						Version:   "",
					}
				} else {
					cl.Container[containerName] = ContainerInfo{
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

func GetPod() ContainerList {
	cl := ContainerList{
		Container: map[string]ContainerInfo{},
	}

	namespaces := []string{
		"kube-system",
	}
	for _, ns := range namespaces {
		resource, err := clientSet.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, d := range resource.Items {
			for _, c := range d.Spec.Containers {
				imageName := c.Image
				containerName := c.Name
				m := regexp.MustCompile("^registry.+net/")
				separateImageRegex := regexp.MustCompile("(.+/)(.+):(.+)")
				rs := separateImageRegex.FindStringSubmatch(imageName)
				if len(rs) < 3 {
					cl.Container[containerName] = ContainerInfo{
						Name:      m.ReplaceAllString(imageName, ""),
						Namespace: ns,
						Registry:  "",
						Image:     imageName,
						Version:   "",
					}
				} else {
					cl.Container[containerName] = ContainerInfo{
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