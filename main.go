package main

import (
	"context"
	"flag"
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"log"
	"os/exec"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	var podName, podNamespace, sideCarName string

	// Passing the flags as below. Filling default values releated to LinkerD
	flag.StringVar(&podName, "podname", "", "Name of the pod") // You can use {{pod.name}} - https://argo-workflows.readthedocs.io/en/latest/variables/#:~:text=Description-,pod.name,-Pod%20name%20of
	flag.StringVar(&podNamespace, "namespace", "", "Namespace where the Argo workflow is executed")
	flag.StringVar(&sideCarName, "sidecar", "linkerd-killer", "Name given to this sidecar in Argo workflow")
	flag.Parse()

	fmt.Printf("LinkerD killer is starting, podname: %s, namespace: %s", podName, podNamespace)

	var config *rest.Config
	var err error
	config, err = rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error building in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}

	ctx := context.Background()

	for {
		pod, err := clientset.CoreV1().Pods(podNamespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				log.Fatalf("Pod %s not found in namespace %s", podName, podNamespace)
			} else {
				log.Fatalf("Error getting pod %s: %v", podName, err)
			}
		}

		// Modify the pod annotation before monitoring the 'wait' container
		err = modifyPodAnnotation(clientset, podNamespace, pod.Name, sideCarName)
		if err != nil {
			log.Fatalf("Error modifying pod annotation: %v", err)
		}

		waitContainerStatus := getContainerStatus(pod, "wait")
		if waitContainerStatus != nil && waitContainerStatus.State.Terminated != nil {
			if waitContainerStatus.State.Terminated.ExitCode == 0 {
				fmt.Println("Wait container completed successfully")
				cmd := exec.Command("curl", "-X", "POST", "http://localhost:4191/shutdown")
				err := cmd.Run()
				if err != nil {
					log.Fatalf("Error executing curl command: %v", err)
				}
				break
			} else {
				fmt.Printf("Wait container exited with non-zero exit code: %d\n", waitContainerStatus.State.Terminated.ExitCode)
				break
			}
		}

		fmt.Println("Wait container is still running, waiting...")
		time.Sleep(5 * time.Second)
	}
}

func getContainerStatus(pod *v1.Pod, containerName string) *v1.ContainerStatus {
	for _, container := range pod.Status.ContainerStatuses {
		if container.Name == containerName {
			return &container
		}
	}
	return nil
}

func modifyPodAnnotation(clientset *kubernetes.Clientset, namespace, podName, sidecar string) error {
	patch := []byte(fmt.Sprintf(`{"metadata": {"annotations": {"workflows.argoproj.io/kill-cmd-linkerd-proxy": "[\"echo\"]", "workflows.argoproj.io/kill-cmd-%s": "[\"echo\"]"}}}`, sidecar))

	_, err := clientset.CoreV1().Pods(namespace).Patch(context.TODO(), podName, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return err
	}

	return nil
}
