/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	//
	// Uncomment to load all auth plugins
	//	"k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

func issueError(w http.ResponseWriter, message string) {
	fmt.Println(message)
	http.Error(w, message, http.StatusBadRequest)
}

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }

///////////////

func router(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	fmt.Println(token)
	if !strings.HasPrefix(token, "Bearer ") {
		issueError(w, "No token")
		return
	}
	companyId := token[7:15]

	subPaths := strings.Split(r.URL.Path, "/")
	service := ""
	if len(subPaths) > 1 {
		service = subPaths[1]
	}

	switch r.Method {
	case "GET":
		list(w, r, companyId)
		return
	case "POST":
		if service == "" {
			issueError(w, "Missing service")
			return
		}
		create(w, r, service, companyId)
		return
	case "DELETE":
		if service == "" {
			issueError(w, "Missing service")
			return
		}
		delete(w, r, service, companyId)
		return
	default:
		issueError(w, "Unsupported mehtod: "+r.Method)
	}
}

// List the deployed services

func list(w http.ResponseWriter, r *http.Request, companyId string) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		issueError(w, err.Error())
		return
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		issueError(w, err.Error())
		return
	}

	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		issueError(w, err.Error())
		return
	}

	// Return the results
	w.Header().Set("Content-Type", "application/json")
	//	services := []string{}
	existingServices := make(map[string]map[string]int)

	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, "tnt-"+companyId+"-") {
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

	fmt.Println(existingServices)
	result, _ := json.Marshal(existingServices)
	w.Write(result)
}

// Create a new service

func create(w http.ResponseWriter, r *http.Request, serviceName string, companyId string) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		issueError(w, err.Error())
		return
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		issueError(w, err.Error())
		return
	}
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	// creates the client
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	name := fmt.Sprintf("tnt-%s-%s", companyId, serviceName)

	// Create the directory and file
	err = os.Mkdir("/tmp/storage/"+name, 0755)

	src := "/tmp/storage/template/handler.py"
	dst := "/tmp/storage/" + name + "/handler.py"
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		issueError(w, "Error: template file not existing")
		return
	}

	if !sourceFileStat.Mode().IsRegular() {
		issueError(w, "Error: template file not a file")
		return
	}

	source, err := os.Open(src)
	if err != nil {
		issueError(w, "Error opening template file")
		return
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		issueError(w, "Error creating default handler file")
		return
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)

	if err != nil {
		issueError(w, err.Error())
		return
	}

	// Create the deployment

	deployment2 := &appsv1.Deployment{
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

	deploymentRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	deployment := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": name,
				"labels": map[string]interface{}{
					"app":                    name,
					"app.kubernetes.io/name": name,
				},
			},
			"spec": map[string]interface{}{
				"replicas": 1,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": name,
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app":                    name,
							"app.kubernetes.io/name": name,
						},
					},

					"spec": map[string]interface{}{
						"containers": []map[string]interface{}{
							{
								"name":  name,
								"image": "svc-python",
								"volumeMounts": []map[string]interface{}{
									{
										"name":      "blockdisk01",
										"mountPath": "/tmp/storage",
										"subPath":   name,
									},
								},
								"ports": []map[string]interface{}{
									{
										"name":          "http",
										"protocol":      "TCP",
										"containerPort": 5000,
									},
								},
							},
						},
						"volumes": []map[string]interface{}{
							{
								"name": "blockdisk01",
								"persistentVolumeClaim": map[string]interface{}{
									"claimName": "pv-main-fs",
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
	result, err := deploymentsClient.Create(context.TODO(), deployment2, metav1.CreateOptions{})
	if 4 != 4 {
		result2, err := client.Resource(deploymentRes).Namespace(apiv1.NamespaceDefault).Create(context.TODO(), deployment, metav1.CreateOptions{})
		if err != nil {
			panic(err)
		}
		fmt.Printf("Created deployment %q.\n", result2.GetName())
	}
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created deployment %q.\n", result.GetName())

	// Service

	serviceClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)

	service2 := &apiv1.Service{
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
	serviceClient.Create(context.TODO(), service2, metav1.CreateOptions{})

	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, "Service created")
}

func delete(w http.ResponseWriter, r *http.Request, serviceName string, companyId string) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		issueError(w, err.Error())
		return
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		issueError(w, err.Error())
		return
	}
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	name := fmt.Sprintf("tnt-%s-%s", companyId, serviceName)

	deletePolicy := metav1.DeletePropagationBackground

	serviceClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)
	err = serviceClient.Delete(context.TODO(), name, metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
	if err != nil {
		issueError(w, err.Error())
		return
	}

	err = deploymentsClient.Delete(context.TODO(), name, metav1.DeleteOptions{
		//		PropagationPolicy:  &deletePolicy,
		GracePeriodSeconds: int64Ptr(0),
	})
	if err != nil {
		issueError(w, err.Error())
		return
	}

	os.RemoveAll("/tmp/storage/" + name)

	fmt.Fprintf(w, "Service deleted")
}

func check(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Health Check</h1>")
}

func main() {
	http.HandleFunc("/", router)
	http.HandleFunc("/health_check", check)
	fmt.Println("Server starting...")
	http.ListenAndServe(":3000", nil)
}
