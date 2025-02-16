package helm

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	yamlSep            = regexp.MustCompile(`(?m)^---`)
	decodingSerializer = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
)

type InstallOptions struct {
	// The namespace to install the chart into
	Namespace string
	// The release name of the chart to install
	ReleaseName string
	// The version of the chart to install
	Version string
	// The name [location] of the chart to install
	ChartName string
	// Values to pass to the chart
	Values map[string]interface{}
	// CreateNamespace if true will create the namespace if it does not exist
	CreateNamespace bool
}

type HelmClient struct {
	settings       *cli.EnvSettings
	registryClient *registry.Client
	log            action.DebugLog
}

// NewInstaller creates a new Helm backed Installer for repo resources.
func NewHelmClient(log action.DebugLog) (*HelmClient, error) {
	if log == nil {
		// If no logger is provided, use a no-op logger.
		log = func(format string, v ...interface{}) {}
	}
	registryClient, err := registry.NewClient(registry.ClientOptEnableCache(true))
	if err != nil {
		log(fmt.Sprintf("failed to create registry client: %v", err))
		return nil, err
	}
	return &HelmClient{
		registryClient: registryClient,
		settings:       cli.New(),
		log:            log,
	}, nil
}

// InstallChart takes a repo's name and a chart name and installs it. If namespace is not empty
// it will install into that namespace and create the namespace. Version is required.
func (h *HelmClient) InstallChart(ctx context.Context, opts InstallOptions) error {
	action, err := h.newInstallAction(opts, false)
	if err != nil {
		return err
	}

	_, err = h.runInstallAction(ctx, opts, action)

	return err
}

func (h *HelmClient) InstallChartCRDs(ctx context.Context, opts InstallOptions, kclient client.Client) ([]*schema.GroupVersionKind, error) {
	action, err := h.newInstallAction(opts, true)
	if err != nil {
		return nil, err
	}

	// Note this isn't actually doing an install, it's equivalent to the `helm template` command
	release, err := h.runInstallAction(ctx, opts, action)
	if err != nil {
		return nil, err
	}

	installedCRDs := []*schema.GroupVersionKind{}
	for _, objYaml := range yamlSep.Split(release.Manifest, -1) {
		crd := &apiextensionsv1.CustomResourceDefinition{}
		_, gvk, err := decodingSerializer.Decode([]byte(objYaml), nil, crd)
		if err != nil {
			h.log(fmt.Sprintf("failed to decode object. Will continue: %v", err))
			continue
		}
		if gvk.Group == "apiextensions.k8s.io" && gvk.Kind == "CustomResourceDefinition" {
			// Make sure we set the annotations expected by helm to 'adopt' them correctly later
			if crd.Annotations == nil {
				crd.Annotations = map[string]string{}
			}
			crd.Annotations["meta.helm.sh/release-name"] = opts.ReleaseName
			crd.Annotations["meta.helm.sh/release-namespace"] = opts.Namespace
			// now server-side apply the CRDs
			err = kclient.Patch(ctx, crd, client.Apply, client.FieldOwner("kubehoist-controller"), client.ForceOwnership)
			if err != nil {
				return nil, fmt.Errorf("failed to apply crd: %w", err)
			}
			for _, version := range crd.Spec.Versions {
				installedCRDs = append(installedCRDs, &schema.GroupVersionKind{Group: crd.Spec.Group, Version: version.Name, Kind: crd.Spec.Names.Kind})
			}
		}
	}

	return installedCRDs, nil
}

func (h *HelmClient) newInstallAction(opts InstallOptions, template bool) (*action.Install, error) {
	actionConfig := &action.Configuration{RegistryClient: h.registryClient}
	if err := actionConfig.Init(h.settings.RESTClientGetter(), opts.Namespace, "", h.log); err != nil {
		return nil, fmt.Errorf("failed to initialize helm action config: %w", err)
	}
	client := action.NewInstall(actionConfig)
	client.Wait = true
	client.Namespace = opts.Namespace
	client.ReleaseName = opts.ReleaseName
	client.Version = opts.Version
	client.CreateNamespace = opts.CreateNamespace
	client.Timeout = 10 * time.Minute
	if !template {
		client.DryRunOption = "none"
	} else {
		client.DryRunOption = "true"
		client.DryRun = true
		client.IncludeCRDs = true
		client.ClientOnly = true
	}

	return client, nil
}

func (h *HelmClient) runInstallAction(ctx context.Context, opts InstallOptions, action *action.Install) (*release.Release, error) {
	chartPath, err := action.ChartPathOptions.LocateChart(opts.ChartName, h.settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart: %w", err)
	}

	ch, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	release, err := action.RunWithContext(ctx, ch, opts.Values)
	if err != nil {
		return nil, fmt.Errorf("failed to install chart: %w", err)
	}

	return release, nil
}
