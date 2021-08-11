package kube

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// New returns a kubernetes client.
// It tries first with in-cluster config, if it fails it will try with out-of-cluster config.
func New() (client kubernetes.Interface, err error) {
	client, err = NewInCluster()
	if err == nil {
		return
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dir = filepath.Join(dir, ".kube", "config")

	client, err = NewFromConfig(dir)
	if err != nil {
		return
	}

	return
}

// NewFromConfig returns a new out-of-cluster kubernetes client.
func NewFromConfig(path string) (client kubernetes.Interface, err error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return
	}

	client = clientset

	return
}

// NewInCluster returns a new in-cluster kubernetes client.
func NewInCluster() (client kubernetes.Interface, err error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return
	}

	client = clientset

	return
}
