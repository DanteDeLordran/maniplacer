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
	Use:   "apply [project-name]",
	Short: "Apply Kubernetes manifests for a project",
	Long:  `Apply all Kubernetes manifests found in the project directory`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		if !utils.IsValidProject() {
			fmt.Printf("Error: Current directory is not a valid Maniplacer project\n")
			os.Exit(1)
		}

		repoName := args[0]

		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			fmt.Printf("Could not get namespace flag, using 'defalt'\n")
			namespace = "default"
		}

		//pick, err := cmd.Flags().GetString("pick")
		//if err != nil {
		//	fmt.Printf("Using latest manifest...\n")
		//}

		if err := initKubeClients(); err != nil {
			fmt.Printf("Error initializing Kubernetes client: %s\n", err)
			os.Exit(1)
		}

		currentPath, err := os.Getwd()
		if err != nil {
			fmt.Printf("Could not get current path: %s\n", err)
			os.Exit(1)
		}

		projectPath := filepath.Join(currentPath, repoName, "manifests", namespace)

		createResources(projectPath, namespace)

	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringP("namespace", "n", "default", "Namespace to apply resources")
	applyCmd.Flags().StringP("pick", "p", "", "Specify a repo manifest version to apply (by default maniplacer applys the latest)")
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

func getLatestManifest(projectPath string) string {

	entries, err := os.ReadDir(projectPath)
	if err != nil {
		fmt.Printf("Could not read dir: %s\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Printf("No manifest versions found in %s\n", projectPath)
		os.Exit(1)
	}

	return filepath.Join(projectPath, entries[len(entries)-1].Name())

}

func createResources(projectPath string, defaultNamespace string) {
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(k8sClient.Discovery()))
	ctx := context.TODO()
	latestManifestPath := getLatestManifest(projectPath)
	entries, err := os.ReadDir(latestManifestPath)
	if err != nil {
		fmt.Printf("Could not read dir: %s\n", err)
		os.Exit(1)
	}

	// Creates the k8s resources found in each entry
	for _, entry := range entries {

		data, err := os.ReadFile(filepath.Join(latestManifestPath, entry.Name()))
		if err != nil {
			fmt.Printf("Could not read file: %s\n", err)
			continue
		}

		obj := &unstructured.Unstructured{}
		err = yaml.Unmarshal(data, &obj.Object)
		if err != nil {
			fmt.Printf("Could not parse YAML: %s\n", err)
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

		existingNamespace, err := k8sClient.CoreV1().Namespaces().Get(ctx, namespace, v1.GetOptions{})
		fmt.Println(existingNamespace)
		if err != nil {
			fmt.Printf("Could not get existing namespace %s\n", err)
		}
		if existingNamespace.Name == "" {
			fmt.Printf("The namespace %s does not exists, do you want to create it? (y/N)\n", namespace)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
				fmt.Printf("namespace '%s' does not exist and creation was declined", namespace)
				continue
			}

			ns := &corev1.Namespace{
				ObjectMeta: v1.ObjectMeta{
					Name: namespace,
					Labels: map[string]string{
						"applier": "maniplacer",
					},
				},
			}

			k8sClient.CoreV1().Namespaces().Create(ctx, ns, v1.CreateOptions{})

		}

		applyOpts := v1.ApplyOptions{FieldManager: "maniplacer"}

		_, err = dynamicClient.Resource(gvr).Namespace(namespace).Apply(context.TODO(), obj.GetName(), obj, applyOpts)
		if err != nil {
			fmt.Printf("apply error: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("%s - Applied!\n", entry.Name())

	}

}
