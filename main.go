package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	NAMESPACE = "default"
	CM_NAME   = "test"
)

func main() {
	mutex := &sync.Mutex{}

	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		panic("Unable to use your kube config, make sure you have a kube config in ~/.kube/config and proper permissions.")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Loading watcher")
	go watchChanges(clientset, NAMESPACE, mutex)
	fmt.Scanln()
}

func watchChanges(clientset *kubernetes.Clientset, namespace string, mutex *sync.Mutex) {
	for {
		watcher, err := clientset.CoreV1().ConfigMaps(NAMESPACE).Watch(
			context.TODO(),
			metav1.SingleObject(
				metav1.ObjectMeta{
					Name:      CM_NAME,
					Namespace: NAMESPACE,
				}))
		if err != nil {
			panic("Watcher issue!")
		}
		fmt.Printf("Watcher::Started watching %v/%v\n", NAMESPACE, CM_NAME)
		onChange(watcher.ResultChan(), mutex)
	}
}

func onChange(eventChan <-chan watch.Event, mutex *sync.Mutex) {
	for {
		event, ok := <-eventChan
		if ok {
			switch event.Type {
			case watch.Added:
				mutex.Lock()
				fmt.Println("Added!")
				mutex.Unlock()
			case watch.Modified:
				mutex.Lock()
				fmt.Println("Modified!")
				mutex.Unlock()
			case watch.Deleted:
				mutex.Lock()
				fmt.Println("Deleted!")
				mutex.Unlock()
			default:
			}
		} else {
			return
		}
	}
}
