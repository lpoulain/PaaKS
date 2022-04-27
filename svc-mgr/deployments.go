package main

import (
	"context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	dep "k8s.io/client-go/kubernetes/typed/apps/v1"
)

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }

func queryRunningDeployments(clientset *kubernetes.Clientset) (map[string]Service, error) {
	/*	// creates the in-cluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		// creates the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, err
		}
	*/
	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	services := make(map[string]Service)

	for _, pod := range pods.Items {
		podNameSubstrings := strings.Split(pod.Name, "-")
		if len(podNameSubstrings) < 3 || podNameSubstrings[0] != "tnt" {
			continue
		}

		key := fmt.Sprintf("tnt-%s-%s", podNameSubstrings[1], podNameSubstrings[2])
		services[key] = Service{
			name:   podNameSubstrings[2],
			tenant: podNameSubstrings[1],
		}
	}

	return services, nil
}

func queryTenantDeployments(tenantCode string, clientset *kubernetes.Clientset) (map[string]map[string]int, error) {
	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	existingServices := make(map[string]map[string]int)

	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, "tnt-"+tenantCode+"-") {
			podNameSubstrings := strings.Split(pod.Name, "-")
			if len(podNameSubstrings) > 2 {
				service := strings.Join(podNameSubstrings[:len(podNameSubstrings)-2], "-")[13:]
				status := string(pod.Status.Phase)
				if pod.ObjectMeta.DeletionTimestamp != nil {
					status = "Terminating"
				}
				if existingServices[service] == nil {
					existingServices[service] = map[string]int{status: 1}
				} else {
					existingServices[service][status] = existingServices[service][status] + 1
				}
			}
		}
	}

	return existingServices, nil
}

func createDeployment(name string, deploymentsClient dep.DeploymentInterface) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app":                    name,
				"app.kubernetes.io/name": name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                    name,
						"app.kubernetes.io/name": name,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            name,
							Image:           "svc-python",
							ImagePullPolicy: "IfNotPresent",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 5000,
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "blockdisk01",
									MountPath: "/tmp/storage",
									SubPath:   name,
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "blockdisk01",
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
									ClaimName: "pv-main-fs",
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Created deployment %q.\n", result.GetName())
	return nil
}

func deleteDeployment(name string, deploymentsClient dep.DeploymentInterface) error {
	err := deploymentsClient.Delete(context.TODO(), name, metav1.DeleteOptions{
		//		PropagationPolicy:  &deletePolicy,
		GracePeriodSeconds: int64Ptr(0),
	})
	return err
}
