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

	"github.com/lpoulain/PaaKS/paaks"

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
	token, err := paaks.GetToken(r)
	if err != nil {
		paaks.IssueError(w, "No token", http.StatusUnauthorized)
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
			paaks.IssueError(w, "Missing service", http.StatusBadRequest)
			return
		}
		create(w, r, service, tenant)
		return
	case "DELETE":
		if service == "" {
			paaks.IssueError(w, "Missing service", http.StatusBadRequest)
			return
		}
		delete(w, r, service, tenant)
		return
	default:
		paaks.IssueError(w, "Unsupported mehtod: "+r.Method, http.StatusBadRequest)
	}
}

// List the deployed services

func list(w http.ResponseWriter, r *http.Request, tenant string) {
	existingServices, err := queryTenantDeployments(tenant[:8], clientset)

	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusInternalServerError)
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
		paaks.IssueError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Create the deployment
	err = createDeployment(serviceFullName, deploymentClient)
	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Service
	err = createTenantService(serviceFullName, clientset)
	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Database
	err = createServiceInDatabase(serviceName, tenant)
	if err != nil {
		paaks.IssueError(w, err.Error(), http.StatusInternalServerError)
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
	definedServices, err := queryServicesInDatabase()
	if err != nil {
		fmt.Println("Error finding services defined in the database", err.Error())
		return
	}

	k8sDeploymentServices, err := queryRunningDeployments(clientset)
	if err == nil {
		for fullServiceName, _ := range definedServices {
			if _, ok := k8sDeploymentServices[fullServiceName]; !ok {
				createDeployment(fullServiceName, deploymentClient)
				fmt.Println("k8s Deployment", fullServiceName, "created")
			}
		}
	}

	k8sServices, err := queryTenantServices(clientset)
	if err == nil {
		for fullServiceName, _ := range definedServices {
			if _, ok := k8sServices[fullServiceName]; !ok {
				createTenantService(fullServiceName, clientset)
				fmt.Println("k8s Service", fullServiceName, "created")
			}
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
