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
	"github.com/k0kubun/pp"
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
	creationTimeStamp time.Time
	State             string
}

type AdminDetails struct {
	email        string
	phone_number string
	name         string
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

type BillingAccount struct {
	UUID             uuid.UUID
	billingAdmins    map[string]AdminDetails
	billingStartDate time.Time
	accountType      AccountType
	balance          float64
	history          map[time.Time]BillFile
	isActive         bool
	company          Company
	projects         []Project
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

type RancherprojectRoleTemplateBindings struct {
	Pagination struct {
		Limit int `json:"limit"`
		Total int `json:"total"`
	} `json:"pagination"`

	Data []struct {
		CreatedTS       int64       `json:"createdTS"`
		CreatorID       interface{} `json:"creatorId"`
		ID              string      `json:"id"`
		Name            string      `json:"name"`
		NamespaceID     interface{} `json:"namespaceId"`
		ProjectID       string      `json:"projectId"`
		RoleTemplateID  string      `json:"roleTemplateId"`
		Type            string      `json:"type"`
		UserID          string      `json:"userId"`
		UserPrincipalID string      `json:"userPrincipalId"`
		UUID            string      `json:"uuid"`
	} `json:"data"`
}

// type namespaceData struct {
// 	Type string `json:"type"`
// 	Data []struct {
// 		ID       string `json:"id"`
// 		Metadata struct {
// 			Annotations struct {
// 				// ProjectId string `json:"field.cattle.io/projectId"`
// 				FieldCattleIoProjectID string `json:"field.cattle.io/projectId"`
// 			} `json:"annotations"`
// 			Labels struct {
// 				FieldCattleIoProjectID   string `json:"field.cattle.io/projectId"`
// 				KubernetesIoMetadataName string `json:"kubernetes.io/metadata.name"`
// 			} `json:"labels"`
// 		} `json:"data"`
// 	}
// }

type namespaceData struct {
	Type  string `json:"type"`
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
	CreateTypes struct {
		Namespace string `json:"namespace"`
	} `json:"createTypes"`
	Actions struct {
	} `json:"actions"`
	ResourceType string `json:"resourceType"`
	Revision     string `json:"revision"`
	Data         []struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Links struct {
			Remove string `json:"remove"`
			Self   string `json:"self"`
			Update string `json:"update"`
			View   string `json:"view"`
		} `json:"links"`
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Metadata   struct {
			Annotations struct {
				CattleIoStatus                       string `json:"cattle.io/status"`
				FieldCattleIoProjectID               string `json:"field.cattle.io/projectId"`
				LifecycleCattleIoCreateNamespaceAuth string `json:"lifecycle.cattle.io/create.namespace-auth"`
				ManagementCattleIoNoDefaultSaToken   string `json:"management.cattle.io/no-default-sa-token"`
				ManagementCattleIoSystemNamespace    string `json:"management.cattle.io/system-namespace"`
			} `json:"annotations"`
			CreationTimestamp string   `json:"creationTimestamp"`
			Fields            []string `json:"fields"`
			Finalizers        []string `json:"finalizers"`
			Labels            struct {
				KubernetesIoMetadataName string `json:"kubernetes.io/metadata.name"`
			} `json:"labels"`
			ManagedFields []struct {
				APIVersion string `json:"apiVersion"`
				FieldsType string `json:"fieldsType"`
				FieldsV1   struct {
					FMetadata struct {
						FAnnotations struct {
							NAMING_FAILED struct {
							} `json:"."`
							FManagementCattleIoSystemNamespace struct {
							} `json:"f:management.cattle.io/system-namespace"`
						} `json:"f:annotations"`
						FLabels struct {
							NAMING_FAILED struct {
							} `json:"."`
							FKubernetesIoMetadataName struct {
							} `json:"f:kubernetes.io/metadata.name"`
						} `json:"f:labels"`
					} `json:"f:metadata"`
				} `json:"fieldsV1"`
				Manager   string `json:"manager"`
				Operation string `json:"operation"`
				Time      string `json:"time"`
			} `json:"managedFields"`
			Name            string      `json:"name"`
			Relationships   interface{} `json:"relationships"`
			ResourceVersion string      `json:"resourceVersion"`
			State           struct {
				Error         bool   `json:"error"`
				Message       string `json:"message"`
				Name          string `json:"name"`
				Transitioning bool   `json:"transitioning"`
			} `json:"state"`
			UID string `json:"uid"`
		} `json:"metadata"`
		Spec struct {
			Finalizers []string `json:"finalizers"`
		} `json:"spec"`
		Status struct {
			Phase string `json:"phase"`
		} `json:"status"`
	} `json:"data"`
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
	// func getRancherUsers() int {
	// Rancher_API, present := os.LookupEnv("Racher_API_Url")
	// if !present {
	// 	panic("Racher_API_Url environment variable is not set!")
	// }
	// return Racher_API, nil

	Rancher_API, RancherBearerToken := getRancherAPIEnvVar()

	//TODO: create a loop to loop over pagination
	url := Rancher_API + "/v3/users"

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

	fmt.Println(RancherResponse)

	return RancherResponse, nil
}

func getRancherProjects() (RancherProjects, error) {
	//TODO: create a loop to loop over pagination
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
	//TODO: create a loop to loop over pagination
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

// this functions return the projects that belong to the inputted user ID
func matchUsersToProjects(userslist RancherUsers) (map[string][]string, error) {

	// map users to an array of their projects
	var usersProjects map[string][]string
	usersProjects = make(map[string][]string)

	Rancher_API, RancherBearerToken := getRancherAPIEnvVar()

	var bearer = "Bearer " + RancherBearerToken

	// // we match projects to their respective users using https://tn.cloud.creometry.com/v3/projectRoleTemplateBindings?userId=${USERID}

	for _, RancherUsersData := range userslist.Data {

		fmt.Println("UserId", RancherUsersData.ID)

		// url := Rancher_API + "/v3/projectRoleTemplateBindings?userId=" + RancherUsersData.ID
		url := Rancher_API + "/v3/projectRoleTemplateBindings"

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
			log.Println("error reading reponse from Rancher API when fetching projectRoleTemplateBindings for user ", RancherUsersData.ID, err)
		}

		RancherResponse := RancherprojectRoleTemplateBindings{}
		jsonErr := json.Unmarshal(body, &RancherResponse)
		if jsonErr != nil {
			log.Fatal(jsonErr)
		}

		// fmt.Println(RancherResponse)
		// fmt.Println("RancherResponse.Data", RancherResponse.Data)

		// userprojects = make(map[string][]string)
		for _, projectRoleTemplateBinding := range RancherResponse.Data {
			if (usersProjects[projectRoleTemplateBinding.UserID]) == nil {
				var projectslist []string
				projectslist = append(projectslist, projectRoleTemplateBinding.ProjectID)
				usersProjects[projectRoleTemplateBinding.UserID] = projectslist

			} else {
				usersProjects[projectRoleTemplateBinding.UserID] = append(usersProjects[projectRoleTemplateBinding.UserID], projectRoleTemplateBinding.ProjectID)
			}

		}

		// for _, namespace := range RancherResponse.Data {

		// 	fmt.Println("namespace.ID", namespace.ID)

		// 	namespacesIds = append(namespacesIds, namespace.ID)

		// }

	}
	fmt.Println("userprojects", usersProjects)
	return usersProjects, nil
}

func bindnamespacestoprojects() (int, error) {
	var projectsNamespaces map[string][]string
	projectsNamespaces = make(map[string][]string)

	_ = projectsNamespaces

	Rancher_API, RancherBearerToken := getRancherAPIEnvVar()

	var bearer = "Bearer " + RancherBearerToken

	url := Rancher_API + "/v1/namespaces"

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
		log.Println("error reading reponse from Rancher API when fetching namespaces from Rancher's API ", err)
	}
	RancherResponse := namespaceData{}
	jsonErr := json.Unmarshal(body, &RancherResponse)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	pp.Print("namespaceData", RancherResponse)

	// for _, namespaceData := range RancherResponse.Data {
	// 	fmt.Println("namespaceData.ID", namespaceData.ID)
	// 	fmt.Println("projectId", namespaceData.Metadata.Annotations.ProjectId)
	// 	projectsNamespaces[namespaceData.Metadata.Annotations.ProjectId] = []string{}
	// 	if (projectsNamespaces[namespaceData.Metadata.Annotations.ProjectId]) == nil {
	// 		var namespacelist []string
	// 		namespacelist = append(namespacelist, namespaceData.ID)
	// 		projectsNamespaces[namespaceData.Metadata.Annotations.ProjectId] = namespacelist

	// 	} else {
	// 		projectsNamespaces[namespaceData.Metadata.Annotations.ProjectId] = append(projectsNamespaces[namespaceData.Metadata.Annotations.ProjectId], namespaceData.ID)
	// 	}
	// }

	// fmt.Println("projectsNamespaces", projectsNamespaces)
	return 0, nil
}

func generatebill() (BillingAccount, error) {
	// TODO: get this objects to be used in writing tests
	adminDetailsexample := make(map[string]AdminDetails)
	adminDetailsexample["admin1"] = AdminDetails{email: "exmaleadmin@email.com", phone_number: "21452012", name: "mohsen"}

	historyExample := make(map[time.Time]BillFile)
	historyExample[time.Date(2022, time.June, 10, 9, 40, 0, 0, time.UTC)] = BillFile{pdfLink: "https://linktopdf.com/12540336", amount: 25}

	projectexample := Project{
		projectId:         uuid.New(),
		clusterId:         uuid.New(),
		creationTimeStamp: time.Now(),
		State:             "Active",
	}
	projectsexample := []Project{projectexample}

	bill := BillingAccount{
		UUID: uuid.New(),
		// get list of billing admins from database or rancher
		billingAdmins: adminDetailsexample,
		// get the first date of any projects in the bill
		billingStartDate: time.Now(),
		// company assigning project if exists
		company:     Company{isCompany: false, TaxId: "", name: ""},
		accountType: "Starter",
		// get balence from database
		balance: 25.410,
		// lsit of previous bills
		history: historyExample,
		// is account suspended or not
		isActive: true,
		// TODO: discuss if this value is better turned to map of clusters and projects whith clusters representing regions
		projects: projectsexample,
	}

	fmt.Println("bill", bill)
	return bill, nil
}

// this functions updates the balence for the accounts on the pay-per-user plan
func updateBalence() {}

// link namesapces to projects and then projects to users

// // we match projects to their respective users using https://tn.cloud.creometry.com/v3/projectRoleTemplateBindings?userId=${USERID}

// // from this one we'll link namespace to their respective projects using data[].metadata.annotations."field.cattle.io/projectId"
func main() {
	LoadDotEnvVariables()

	// RancherUsersDetails, err := getRancherUsers()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// matchUsersToProjects(RancherUsersDetails)

	bindnamespacestoprojects()

	// generatebill()
	// clusterIds := getRancherClustersIds()

	// getRancherNamespaces(clusterIds)

	// // getRancherProjects()
	// // var namespaces = []string{"default"}
	// namespaceMetrics, err := getNamespaceMetrics()
	// if err != nil {
	// 	log.Println("error geeting namespace metrics from Kubecost's API", err)
	// }

	// fmt.Println("namespaceMetrics", namespaceMetrics)

}
