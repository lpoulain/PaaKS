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
	"net/http"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func error(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusBadRequest)
}

func list(w http.ResponseWriter, r *http.Request) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		error(w, err.Error())
		return
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		error(w, err.Error())
		return
	}

	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		error(w, err.Error())
		return
	}

	// Return the results
	w.Header().Set("Content-Type", "application/json")
	services := []string{}
	existingServices := make(map[string]bool)
	token := r.Header.Get("Authorization")
	fmt.Println(token)
	if !strings.HasPrefix(token, "Bearer ") {
		fmt.Fprintf(w, "No token")
		return
	}

	companyId := token[7:15]

	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, "tnt-"+companyId+"-") {
			podNameSubstrings := strings.Split(pod.Name, "-")
			if len(podNameSubstrings) > 2 {
				service := strings.Join(podNameSubstrings[:len(podNameSubstrings)-2], "-")[13:]
				if !existingServices[service] {
					services = append(services, service)
					existingServices[service] = true
				}
			}
		}
	}

	fmt.Println(services)
	result, _ := json.Marshal(services)
	w.Write(result)
}

func check(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Health Check</h1>")
}

func main() {
	http.HandleFunc("/", list)
	http.HandleFunc("/health_check", check)
	fmt.Println("Server starting...")
	http.ListenAndServe(":3000", nil)
}
