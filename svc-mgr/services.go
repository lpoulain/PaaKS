package main

import (
	"context"
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

func queryTenantServices(clientset *kubernetes.Clientset) (map[string]Service, error) {
	serviceClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)

	services := make(map[string]Service)

	svcs, err := serviceClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, svc := range svcs.Items {
		podNameSubstrings := strings.Split(svc.Name, "-")
		if len(podNameSubstrings) != 3 || podNameSubstrings[0] != "tnt" {
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

func createTenantService(name string, clientset *kubernetes.Clientset) error {
	serviceClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)

	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Port: 80,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 5000,
					},
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/name": name,
			},
			ClusterIP: "",
		},
	}

	fmt.Println("Creating service...")
	_, err := serviceClient.Create(context.TODO(), service, metav1.CreateOptions{})

	if err != nil {
		return nil
	}

	fmt.Println("Service created")

	return nil
}

func deleteTenantService(name string, clientset *kubernetes.Clientset) error {
	deletePolicy := metav1.DeletePropagationBackground

	serviceClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)
	err := serviceClient.Delete(context.TODO(), name, metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
	return err
}
