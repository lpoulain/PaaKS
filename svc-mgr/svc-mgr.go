package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	dep "k8s.io/client-go/kubernetes/typed/apps/v1"

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

var clientset *kubernetes.Clientset
var deploymentClient dep.DeploymentInterface

func issueError(w http.ResponseWriter, message string) {
	fmt.Println(message)
	http.Error(w, message, http.StatusBadRequest)
}

type Service struct {
	name   string
	tenant string
}

func getK8sResources() (*kubernetes.Clientset, dep.DeploymentInterface, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	return clientset, deploymentsClient, nil
}

///////////////

func router(w http.ResponseWriter, r *http.Request) {
	token, err := getToken(r)
	if err != nil {
		issueError(w, "No token")
		return
	}
	tenant := token.Tenant

	subPaths := strings.Split(r.URL.Path, "/")
	service := ""
	if len(subPaths) > 1 {
		service = subPaths[1]
	}

	switch r.Method {
	case "GET":
		list(w, r, tenant)
		return
	case "POST":
		if service == "" {
			issueError(w, "Missing service")
			return
		}
		create(w, r, service, tenant)
		return
	case "DELETE":
		if service == "" {
			issueError(w, "Missing service")
			return
		}
		delete(w, r, service, tenant)
		return
	default:
		issueError(w, "Unsupported mehtod: "+r.Method)
	}
}

// List the deployed services

func list(w http.ResponseWriter, r *http.Request, tenant string) {
	existingServices, err := queryTenantDeployments(tenant[:8], clientset)

	if err != nil {
		issueError(w, err.Error())
		return
	}

	// Return the results
	w.Header().Set("Content-Type", "application/json")
	fmt.Println(existingServices)
	result, _ := json.Marshal(existingServices)
	w.Write(result)
}

// Create a new service

func create(w http.ResponseWriter, r *http.Request, serviceName string, tenant string) {
	serviceFullName := fmt.Sprintf("tnt-%s-%s", tenant[:8], serviceName)

	// Create the directory and file
	err := createServiceFiles(serviceFullName)
	if err != nil {
		issueError(w, err.Error())
		return
	}
	// Create the deployment
	err = createDeployment(serviceFullName, deploymentClient)
	if err != nil {
		issueError(w, err.Error())
		return
	}
	// Service
	err = createTenantService(serviceFullName, clientset)
	if err != nil {
		issueError(w, err.Error())
		return
	}
	// Database
	err = createServiceInDatabase(serviceName, tenant)
	if err != nil {
		issueError(w, err.Error())
	}

	fmt.Fprintf(w, "Service created")
}

func delete(w http.ResponseWriter, r *http.Request, serviceName string, tenant string) {
	serviceFullName := fmt.Sprintf("tnt-%s-%s", tenant[:8], serviceName)

	deleteServiceInDatabase(serviceName, tenant)
	deleteTenantService(serviceFullName, clientset)
	deleteDeployment(serviceFullName, deploymentClient)
	deleteServiceFiles(serviceFullName)

	fmt.Fprintf(w, "Service deleted")
}

func createMissingDeployments() {
	runningServices, err := queryRunningDeployments(clientset)
	if err != nil {
		fmt.Println("Error finding running deployments", err.Error())
		return
	}

	definedServices, err := queryServicesInDatabase()

	for fullServiceName, _ := range definedServices {
		if _, ok := runningServices[fullServiceName]; !ok {
			createDeployment(fullServiceName, deploymentClient)
			createTenantService(fullServiceName, clientset)
		}
	}
}

func check(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Health Check</h1>")
}

func main() {
	c, d, err := getK8sResources()

	clientset = c
	deploymentClient = d

	if err != nil {
		fmt.Println("Error getting Kubernetes resources", err.Error())
	} else {
		createMissingDeployments()
	}

	http.HandleFunc("/", router)
	http.HandleFunc("/health_check", check)
	fmt.Println("Server starting...")
	http.ListenAndServe(":3000", nil)
}
