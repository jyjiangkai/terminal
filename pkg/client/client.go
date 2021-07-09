package client

import (
	"fmt"
	"k8s.io/client-go/dynamic"
	"sync"
	"time"

	"terminal/utils"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeConfigPath = utils.Env("KUBE_CONFIG_PATH", "config")

var defaultDuration = time.Duration(time.Second * 5)
var kubeClientset kubernetes.Interface
var dynamicClientset dynamic.Interface
var kubeconfig *rest.Config
var once sync.Once

func init() {
	fmt.Println("initializing...")
	BuildClientset()
}

func setupKubeClientset() error {
	c, err := NewKubeInClusterClient()
	if err == nil {
		fmt.Printf("using in cluster clientset...\n")
		kubeClientset = c
		return nil
	}

	config, err := utils.ReadFile(kubeConfigPath)
	if err != nil {
		fmt.Printf("cannot read kube config file, err: %v \n", err)
		return err
	}
	c, err = NewKubeOutClusterClient(config)
	if err != nil {
		fmt.Printf("cannot create kube out cluster clientset, err: %v \n", err)
		return err
	}
	fmt.Printf("using out cluster clientset...\n")
	kubeClientset = c
	return nil
}

func setupDynamicClientset() error {
	c, err := NewDynamicInClusterClient()
	if err == nil {
		fmt.Printf("using in cluster clientset...\n")
		dynamicClientset = c
		return nil
	}

	config, err := utils.ReadFile(kubeConfigPath)
	if err != nil {
		fmt.Printf("cannot read dynamic config file, err: %v \n", err)
		return err
	}
	c, err = NewDynamicOutClusterClient(config)
	if err != nil {
		fmt.Printf("cannot create dynamic out cluster clientset, err: %v \n", err)
		return err
	}
	fmt.Printf("using out cluster clientset...\n")
	dynamicClientset = c

	return nil
}

// BuildClientset build cache factory and start informers
func BuildClientset() {
	once.Do(func() {
		if err := setupKubeClientset(); err != nil {
			panic(err)
		}
		if err := setupDynamicClientset(); err != nil {
			panic(err)
		}
	})
}

// KubeClientset return clientset
func KubeClientset() kubernetes.Interface {
	return kubeClientset
}

// DynamicClientset return clientset
func DynamicClientset() dynamic.Interface {
	return dynamicClientset
}

// NewDynamicInClusterClient
func NewDynamicInClusterClient() (dynamic.Interface, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

// NewKubeOutClusterClient creates a out cluster kubernetes clientset interface
func NewDynamicOutClusterClient(config []byte) (dynamic.Interface, error) {
	cfg, err := LoadKubeConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize inclusterconfig: %v", err)
	}
	clientset, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize client: %v", err)
	}
	return clientset, nil
}

// NewKubeInClusterClient creates an in cluster kubernetes clientset interface
func NewKubeInClusterClient() (kubernetes.Interface, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to initialize inclusterconfig: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize client: %v", err)
	}
	return clientset, nil
}

// NewKubeOutClusterClient creates a out cluster kubernetes clientset interface
func NewKubeOutClusterClient(config []byte) (kubernetes.Interface, error) {
	cfg, err := LoadKubeConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize inclusterconfig: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize client: %v", err)
	}
	return clientset, nil
}

// NewKubeClientWithConfigPath creates a out cluster kubernetes clientset interface
func NewKubeClientWithConfigPath(configPath string) (kubernetes.Interface, error) {
	config, err := utils.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read kube config file: %v", err)
	}
	cfg, err := LoadKubeConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize inclusterconfig: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize client: %v", err)
	}
	return clientset, nil
}

func LoadKubeConfig(config []byte) (*rest.Config, error) {
	c, err := clientcmd.Load(config)
	if err != nil {
		return nil, err
	}
	clientConfig := clientcmd.NewDefaultClientConfig(*c, &clientcmd.ConfigOverrides{})
	return clientConfig.ClientConfig()
}

// NewSharedInformerFactory creates a new SharedInformerFactory
func NewSharedInformerFactory(clientset kubernetes.Interface) (informers.SharedInformerFactory, error) {
	sharedInformers := informers.NewSharedInformerFactory(clientset, defaultDuration)
	return sharedInformers, nil
}

// Config get kube config
func Config() (*rest.Config, error) {
	if kubeconfig != nil {
		return kubeconfig, nil
	}
	cfg, err := rest.InClusterConfig()
	if err == nil {
		kubeconfig = nil
		return cfg, nil
	}
	config, err := utils.ReadFile(kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read kube config file: %v", err)
	}
	cfg, err = LoadKubeConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get inclusterconfig: %v", err)
	}
	kubeconfig = cfg
	return kubeconfig, nil
}
