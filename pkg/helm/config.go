package helm

import (
	"os"

	newkube "github.com/alauda/captain/pkg/kube"
	"github.com/alauda/captain/pkg/kubeconfig"
	releaseclient "github.com/alauda/helm-crds/pkg/client/clientset/versioned"
	"helm.sh/helm/pkg/kube"

	"github.com/alauda/captain/pkg/release/storagedriver"
	"helm.sh/helm/pkg/action"
	"helm.sh/helm/pkg/storage"
	"helm.sh/helm/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog"
)

// getNamespace get the namespaces from ..... This may be a little unnecessary, may be we can just
// use the one we know.
func getNamespace(flags *genericclioptions.ConfigFlags) string {
	if ns, _, err := flags.ToRawKubeConfigLoader().Namespace(); err == nil {
		return ns
	}
	return "alauda-system"
}

// newActionConfig create a config for all the actions(install,delete,update...)
// allNamespaces is always set to false for now,
// default storage driver is Release now
func (d *Deploy) newActionConfig() (*action.Configuration, error) {
	cfg, err := kubeconfig.UpdateKubeConfig(d.Cluster)
	if err != nil {
		return nil, err
	}
	cfg.Namespace = d.Cluster.Namespace

	cfgFlags := kube.GetConfig(cfg.Path, cfg.Context, cfg.Namespace)
	kc := newkube.New(cfgFlags)
	// hope it works
	kc.Log = klog.Infof

	namespace := getNamespace(cfgFlags)

	relClientSet, err := releaseclient.NewForConfig(d.Cluster.ToRestConfig())
	if err != nil {
		return nil, err
	}

	var store *storage.Storage
	switch os.Getenv("HELM_DRIVER") {
	case "release", "releases", "":
		d := storagedriver.NewReleases(relClientSet.AppV1alpha1().Releases(namespace))
		d.Log = klog.Infof
		store = storage.Init(d)
	case "memory":
		d := driver.NewMemory()
		store = storage.Init(d)
	default:
		// Not sure what to do here.
		panic("Unknown driver in HELM_DRIVER: " + os.Getenv("HELM_DRIVER"))
	}

	return &action.Configuration{
		RESTClientGetter: cfgFlags,
		KubeClient:       kc,
		Releases:         store,
		Log:              klog.Infof,
	}, nil
}
