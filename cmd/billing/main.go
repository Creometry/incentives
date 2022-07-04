package main

// env vars
// APP_ENV = "development" || "production"
// Racher_API = "link-to-rancher-api"

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/kubecost/opencost/pkg/kubecost"
)

type AccountType string

const (
	PayPerUse string = "PayPerUse"
	Starter          = "Starter"
	Pro              = "Pro"
	Elite            = "Elite"
)

type Project struct {
	projectId         uuid.UUID
	clusterId         uuid.UUID
	creationTimeStamp string
}

type AdminDetails struct {
	email        string
	phone_number string
}

type Company struct {
	isCompany bool
	TaxId     string
	name      string
}

type BillFile struct {
	pdfLink string
	amount  float64
}

type Bill struct {
	UUID          uuid.UUID
	billingAdmins map[string]AdminDetails
	StartDate     time.Time
	company       Company
	accountType   AccountType
	balance       float64
	history       map[time.Time]BillFile
	isActive      bool
	regions       []Project
}

type resourcePricing struct {
	CpuCoreUsageMinuteBilling float64
	RamByteUsageMinuteBilling float64
}

type Metrics struct {
	CpuMinutes      int64
	CpuAverageUsage float64
	RamMinutes      int64
	RamAverageUsage float64
}

type allocationResponse struct {
	Code int                              `json:"code"`
	Data []map[string]kubecost.Allocation `json:"data"`
}

type RancherUsers struct {
	Pagination struct {
		Limit int `json:"limit"`
		Total int `json:"total"`
	} `json:"pagination"`

	Data []struct {
		CreatedTS int64  `json:"createdTS"`
		CreatorID string `json:"creatorId"`
		Enabled   bool   `json:"enabled"`
		ID        string `json:"id"`
		Username  string `json:"username,omitempty"`
		UUID      string `json:"uuid"`
	}
}

type RancherProjects struct {
	Pagination struct {
		Limit int `json:"limit"`
		Total int `json:"total"`
	} `json:"pagination"`

	Data []struct {
		ClusterID string    `json:"clusterId"`
		Created   time.Time `json:"created"`
		CreatedTS int64     `json:"createdTS"`
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		UUID      string    `json:"uuid"`
	}
}

type RancheclusterId struct {
	Data []struct {
		ID string `json:"id"`
	}
}

type RancherNamespaces struct {
	Data []struct {
		ID string `json:"id"`
	}
}

func LoadDotEnvVariables() int {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return 0
}

func getresourcePricing() resourcePricing {
	//TODO: get resource pricing from API
	resourcePricingTestValues := resourcePricing{
		CpuCoreUsageMinuteBilling: 0.04,
		RamByteUsageMinuteBilling: 0.00006,
	}

	return resourcePricingTestValues
}

// func getNamespaceMetrics(namespaces []string) (allocationResponse, error) {
func getNamespaceMetrics() (allocationResponse, error) {
	var kubecostUrl string
	// ** kubecost allocation params: **
	// ref : https://github.com/kubecost/docs/blob/main/allocation.md

	// namespace of the user to be billed:

	// TODO: put namespaces function variable here
	// Namespaces := "default"

	// window of billed time
	window := "month"
	// accumulate results (kubecost returns dates seperated by day if this is false (default value) )
	accumulte := "true"
	//field by wich to aggrgate results
	aggregate := "namespace"

	// kubecostUrl := "kubecost-cost-analyzer"
	if os.Getenv("APP_ENV") == "development" {
		kubecostUrl = "localhost"
	} else {
		kubecostUrl = "kubecost-cost-analyzer"
	}

	// kubecost metrics api
	// url := "http://" + kubecostUrl + ":9090/model/allocation?window=" + window + "&accumulate=" + accumulte + "&filterNamespaces=" + Namespaces + "&aggregate=" + aggregate
	url := "http://" + kubecostUrl + ":9090/model/allocation?window=" + window + "&accumulate=" + accumulte + "&aggregate=" + aggregate

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("No response from Kubecost!")
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(body))

	kubecostResponse := allocationResponse{}
	jsonErr := json.Unmarshal(body, &kubecostResponse)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	// fmt.Println(kubecostResponse)
	return kubecostResponse, nil

}

// TODO: write function that gets node metrics
// func getNodeMetrics(namespaces []string) (allocationResponse, error) {}

func getRancherAPIEnvVar() (string, string) {
	Rancher_API, present := os.LookupEnv("Rancher_API_Url")
	if !present {
		panic("Rancher_API_Url environment variable is not set!")
	}

	//TODO: catch 401 unauthorized token error
	RancherBearerToken, bearerpresent := os.LookupEnv("RancherBearerToken")
	if !bearerpresent {
		panic("RancherBearerToken environment variable is not set!")
	}

	return Rancher_API, RancherBearerToken
}

func getRancherUsers() (RancherUsers, error) {
	// Rancher_API, present := os.LookupEnv("Racher_API_Url")
	// if !present {
	// 	panic("Racher_API_Url environment variable is not set!")
	// }
	// return Racher_API, nil

	Rancher_API, RancherBearerToken := getRancherAPIEnvVar()

	//TODO: create a loop to loop over pagination
	url := Rancher_API + "/users"

	var bearer = "Bearer " + RancherBearerToken

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("No response from Rancher!", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("error reading reponse from Rancher API!", err)
	}

	RancherResponse := RancherUsers{}
	jsonErr := json.Unmarshal(body, &RancherResponse)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	// fmt.Println(RancherResponse)

	//TODO: write code that maps Rancher's id to username
	return RancherResponse, nil

}

func getRancherProjects() (RancherProjects, error) {
	Rancher_API, RancherBearerToken := getRancherAPIEnvVar()

	url := Rancher_API + "/projects"

	var bearer = "Bearer " + RancherBearerToken

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("No response from Rancher!", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("error reading reponse from Rancher API!", err)
	}

	RancherResponse := RancherProjects{}
	jsonErr := json.Unmarshal(body, &RancherResponse)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	fmt.Println(RancherResponse)

	return RancherResponse, nil
}

func getRancherClustersIds() RancheclusterId {
	Rancher_API, RancherBearerToken := getRancherAPIEnvVar()

	url := Rancher_API + "/v3/clusters"

	var bearer = "Bearer " + RancherBearerToken

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("No response from Rancher!", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("error reading reponse from Rancher API!", err)
	}

	RancherResponse := RancheclusterId{}
	jsonErr := json.Unmarshal(body, &RancherResponse)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	fmt.Println(RancherResponse)

	return RancherResponse
}

// func gerRancherNamespaces(clusterList RancheclusterId) (RancherNamespaces, error) {
func getRancherNamespaces(clusterList RancheclusterId) (int, error) {

	var namespacesIds []string

	Rancher_API, RancherBearerToken := getRancherAPIEnvVar()

	var bearer = "Bearer " + RancherBearerToken

	for _, cluster := range clusterList.Data {
		fmt.Println("clusterId", cluster)

		url := Rancher_API + "/k8s/clusters/" + cluster.ID + "/v1/namespaces"

		req, err := http.NewRequest("GET", url, nil)
		req.Header.Add("Authorization", bearer)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("No response from Rancher!", err)
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("error reading reponse from Rancher API when fetching namespaces for cluster", cluster.ID, err)
		}

		RancherResponse := RancherNamespaces{}
		jsonErr := json.Unmarshal(body, &RancherResponse)
		if jsonErr != nil {
			log.Fatal(jsonErr)
		}

		// fmt.Println(RancherResponse)
		// fmt.Println("RancherResponse.Data", RancherResponse.Data)

		for _, namespace := range RancherResponse.Data {

			fmt.Println("namespace.ID", namespace.ID)

			namespacesIds = append(namespacesIds, namespace.ID)

		}

		fmt.Println("clusterId", cluster)
	}
	return 0, nil
}

// link namesapces to projects and then projects to users
func main() {
	LoadDotEnvVariables()
	clusterIds := getRancherClustersIds()
	// we link projects to their users using https://tn.cloud.creometry.com/v3/projectRoleTemplateBindings?userId=${USERID}

	// from this one we'll link namespace to their respective projects using data[].metadata.annotations."field.cattle.io/projectId"

	getRancherNamespaces(clusterIds)
	// getRancherUsers()
	// getRancherProjects()
	// var namespaces = []string{"default"}
	namespaceMetrics, err := getNamespaceMetrics()
	if err != nil {
		log.Println("error geeting namespace metrics from Kubecost's API", err)
	}

	fmt.Println("namespaceMetrics", namespaceMetrics)

}
