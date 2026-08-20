package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
	Long:  `Apply the latest generated Kubernetes manifests for a repository.`,
	Args:  cobra.ExactArgs(1),
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home dir: %w", err)
	}

	kubeconfig := filepath.Join(homeDir, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
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
		if entry.IsDir() {
			latest = entry.Name()
		}
	}

	if latest == "" {
		return "", fmt.Errorf("no manifest versions found in %s", projectPath)
	}

	return filepath.Join(projectPath, latest), nil
}

func createResources(projectPath string, defaultNamespace string) error {
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

		obj := &unstructured.Unstructured{}
		err = yaml.Unmarshal(data, &obj.Object)
		if err != nil {
			fmt.Printf("Could not parse YAML: %s\n", err)
			errorCount++
			continue
		}

		// Skip empty documents
		if obj.GetKind() == "" {
			continue
		}

		gvk := obj.GroupVersionKind()

		restMapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			fmt.Printf("Could not create rest mapper: %s\n", err)
			errorCount++
			continue
		}

		gvr := restMapping.Resource

		namespace := obj.GetNamespace()
		if namespace == "" {
			// Check if this resource is namespaced
			if restMapping.Scope.Name() == "namespace" {
				namespace = defaultNamespace
				obj.SetNamespace(namespace)
			}
		}

		if namespace != "" {
			if err := ensureNamespace(ctx, namespace); err != nil {
				fmt.Printf("Skipping %s: %s\n", entry.Name(), err)
				errorCount++
				continue
			}
		}

		applyOpts := v1.ApplyOptions{FieldManager: "maniplacer"}

		_, err = dynamicClient.Resource(gvr).Namespace(namespace).Apply(ctx, obj.GetName(), obj, applyOpts)
		if err != nil {
			fmt.Printf("Could not apply %s: %s\n", entry.Name(), err)
			errorCount++
			continue
		}

		fmt.Printf("%s - Applied!\n", entry.Name())
		appliedCount++
	}

	fmt.Printf("\nApply complete: %d applied, %d errors\n", appliedCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("apply completed with %d errors", errorCount)
	}

	return nil
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
