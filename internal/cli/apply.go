package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/dantedelordran/maniplacer/internal/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	k8sClient     *kubernetes.Clientset
	dynamicClient dynamic.Interface
)

var applyCmd = &cobra.Command{
	Use:   "apply [repository]",
	Short: "Apply generated Kubernetes manifests for a repository",
	Long: `Apply the latest generated Kubernetes manifests for a repository.

The target cluster is the current context of the KUBECONFIG environment variable, falling back to ~/.kube/config.

Every YAML document in each manifest is applied. A manifest whose metadata.namespace differs from --namespace is rejected rather than applied to that other namespace.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !utils.IsValidProject() {
			return fmt.Errorf("current directory is not a valid Maniplacer project")
		}

		repoName := args[0]
		if err := utils.ValidateRepoName(repoName); err != nil {
			return fmt.Errorf("invalid repository name: %w", err)
		}

		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			return fmt.Errorf("could not get namespace flag: %w", err)
		}
		if err := utils.ValidateNamespace(namespace); err != nil {
			return fmt.Errorf("invalid namespace: %w", err)
		}

		if err := initKubeClients(); err != nil {
			return fmt.Errorf("error initializing Kubernetes client: %w", err)
		}

		currentPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not get current path: %w", err)
		}

		projectPath := filepath.Join(currentPath, repoName, "manifests", namespace)

		if err := createResources(projectPath, namespace); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringP("namespace", "n", utils.DefaultNamespace, "Namespace to apply resources")
}

func initKubeClients() error {
	// Matches kubectl: KUBECONFIG when set, otherwise ~/.kube/config
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return fmt.Errorf("could not build kubeconfig: %w", err)
	}

	k8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("could not create kubernetes client: %w", err)
	}

	dynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("could not create dynamic client: %w", err)
	}

	return nil
}

// getLatestManifest returns the most recent timestamped manifest directory.
// ReadDir sorts by name and the timestamp format sorts lexicographically in
// chronological order, so the last directory entry is the newest one.
func getLatestManifest(projectPath string) (string, error) {
	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return "", fmt.Errorf("could not read manifests directory: %w", err)
	}

	latest := ""
	for _, entry := range entries {
		if _, err := time.Parse(versionLayout, entry.Name()); entry.IsDir() && err == nil {
			latest = entry.Name()
		}
	}

	if latest == "" {
		return "", fmt.Errorf("no manifest versions found in %s", projectPath)
	}

	return filepath.Join(projectPath, latest), nil
}

// decodeManifests parses every document in a YAML stream, skipping empty ones.
func decodeManifests(data []byte) ([]*unstructured.Unstructured, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	var objects []*unstructured.Unstructured
	for {
		obj := &unstructured.Unstructured{}
		err := decoder.Decode(&obj.Object)
		if errors.Is(err, io.EOF) {
			return objects, nil
		}
		if err != nil {
			return nil, err
		}
		if len(obj.Object) > 0 {
			objects = append(objects, obj)
		}
	}
}

// checkManifestNamespace rejects a manifest that sets a namespace other than the
// target. One without a namespace is fine: it is either cluster-scoped or placed
// in the target by apply.
func checkManifestNamespace(obj *unstructured.Unstructured, target string) error {
	if namespace := obj.GetNamespace(); namespace != "" && namespace != target {
		return fmt.Errorf("%s '%s' sets namespace '%s', which does not match target namespace '%s'", obj.GetKind(), obj.GetName(), namespace, target)
	}
	return nil
}

func createResources(projectPath string, targetNamespace string) error {
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(k8sClient.Discovery()))
	ctx := context.TODO()

	latestManifestPath, err := getLatestManifest(projectPath)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(latestManifestPath)
	if err != nil {
		return fmt.Errorf("could not read manifest directory: %w", err)
	}

	appliedCount := 0
	errorCount := 0

	// Creates the k8s resources found in each entry
	for _, entry := range entries {

		data, err := os.ReadFile(filepath.Join(latestManifestPath, entry.Name()))
		if err != nil {
			fmt.Printf("Could not read file: %s\n", err)
			errorCount++
			continue
		}

		objects, err := decodeManifests(data)
		if err != nil {
			fmt.Printf("Could not parse YAML in %s: %s\n", entry.Name(), err)
			errorCount++
			continue
		}

		for _, obj := range objects {
			if err := applyObject(ctx, mapper, obj, targetNamespace); err != nil {
				fmt.Printf("Could not apply %s: %s\n", entry.Name(), err)
				errorCount++
				continue
			}

			fmt.Printf("%s - Applied %s '%s'!\n", entry.Name(), obj.GetKind(), obj.GetName())
			appliedCount++
		}
	}

	fmt.Printf("\nApply complete: %d applied, %d errors\n", appliedCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("apply completed with %d errors", errorCount)
	}

	return nil
}

// applyObject server-side applies a single resource to the target namespace.
func applyObject(ctx context.Context, mapper *restmapper.DeferredDiscoveryRESTMapper, obj *unstructured.Unstructured, targetNamespace string) error {
	if err := checkManifestNamespace(obj, targetNamespace); err != nil {
		return err
	}

	gvk := obj.GroupVersionKind()

	restMapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("could not create rest mapper: %w", err)
	}

	gvr := restMapping.Resource

	namespace := obj.GetNamespace()
	if namespace == "" {
		// Check if this resource is namespaced
		if restMapping.Scope.Name() == "namespace" {
			namespace = targetNamespace
			obj.SetNamespace(namespace)
		}
	}

	if namespace != "" {
		if err := ensureNamespace(ctx, namespace); err != nil {
			return err
		}
	}

	applyOpts := v1.ApplyOptions{FieldManager: "maniplacer"}

	_, err = dynamicClient.Resource(gvr).Namespace(namespace).Apply(ctx, obj.GetName(), obj, applyOpts)
	return err
}

// ensureNamespace checks that the target namespace exists, offering to create it
// if it does not.
func ensureNamespace(ctx context.Context, namespace string) error {
	_, err := k8sClient.CoreV1().Namespaces().Get(ctx, namespace, v1.GetOptions{})
	if err == nil {
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("could not check namespace '%s': %w", namespace, err)
	}

	if !utils.ConfirmMessage(fmt.Sprintf("The namespace '%s' does not exist, do you want to create it?", namespace)) {
		return fmt.Errorf("namespace '%s' does not exist and creation was declined", namespace)
	}

	ns := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"applier": "maniplacer",
			},
		},
	}

	if _, err := k8sClient.CoreV1().Namespaces().Create(ctx, ns, v1.CreateOptions{}); err != nil {
		return fmt.Errorf("could not create namespace '%s': %w", namespace, err)
	}

	return nil
}
