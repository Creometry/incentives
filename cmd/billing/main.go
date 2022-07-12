package main

// env vars
// APP_ENV = "development" || "production"
// Racher_API = "link-to-rancher-api"

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/k0kubun/pp"
	"github.com/kubecost/opencost/pkg/kubecost"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type AccountType string

const (
	PayPerUse string = "PayPerUse"
	Starter          = "Starter"
	Pro              = "Pro"
	Elite            = "Elite"
)

type Project struct {
	projectId          uuid.UUID `gorm:"primaryKey"`
	clusterId          uuid.UUID
	creationTimeStamp  time.Time
	State              string
	BillingAccountUUID string `gorm:"foreignKey:BillingAccountRefer"`
}

type AdminDetails struct {
	gorm.Model
	UUID               uuid.UUID `json:"uuid"`
	Email              string    `json:"email"`
	Phone_number       string    `json:"phone_number"`
	Name               string    `json:"name"`
	BillingAccountUUID string    `gorm:"foreignKey:BillingAccountRefer"`
}

type Company struct {
	IsCompany bool   `json:"isCompany"`
	TaxId     string `json:"TaxId" gorm:"primaryKey"`
	Name      string `json:"name"`
}

type BillFile struct {
	BillingDate        time.Time `json:"BillingDate" gorm:"primaryKey"`
	PdfLink            string    `json:"pdfLink"`
	Amount             float64   `json:"amount"`
	BillingAccountUUID string    `gorm:"foreignKey:BillingAccountRefer"`
}

type BillingAccount struct {
	gorm.Model
	UUID uuid.UUID `json:"uuid" gorm:"primaryKey"`
	// BillingAdmins []AdminDetails `json:"billingAdmins" gorm:"embedded"`
	BillingAdmins []AdminDetails `json:"billingAdmins"`
	// BillingAdmins    []AdminDetails `json:"billingAdmins" gorm:"many2many:AdminDetails;"`
	BillingStartDate time.Time   `json:"billingStartDate"`
	AccountType      AccountType `json:"accountType"`
	Balance          float64     `json:"balance"`
	History          []BillFile  `json:"history"`
	IsActive         bool        `json:"isActive"`
	// Company          Company        `json:"company" gorm:"references:TaxId"`
	Company  Company   `json:"company" gorm:"-"`
	Projects []Project `json:"projects"`
}

type resourcePricing struct {
	CpuCoreUsageMinuteBilling float64
	RamByteUsageMinuteBilling float64
}

type Metrics struct {
	CPUCoreHours         float64
	CpuAverageUsage      float64
	RamMinutes           float64
	RamAverageUsage      float64
	networkTransferBytes float64
	networkReceiveBytes  float64
	pvByteHours          float64
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
func getNamespaceMetrics(namespaceId string) (allocationResponse, error) {
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
	// select namespace by which to filter results
	filterNamespaces := namespaceId

	// _ = filterNamespaces

	// kubecostUrl := "kubecost-cost-analyzer"
	if os.Getenv("APP_ENV") == "development" {
		kubecostUrl = "localhost"
	} else {
		kubecostUrl = "kubecost-cost-analyzer"
	}

	// kubecost metrics api
	// url := "http://" + kubecostUrl + ":9090/model/allocation?window=" + window + "&accumulate=" + accumulte + "&filterNamespaces=" + Namespaces + "&aggregate=" + aggregate
	url := "http://" + kubecostUrl + ":9090/model/allocation?window=" + window + "&accumulate=" + accumulte + "&aggregate=" + aggregate + "&filterNamespaces=" + filterNamespaces

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("No response from Kubecost!")
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	kubecostResponse := allocationResponse{}
	jsonErr := json.Unmarshal(body, &kubecostResponse)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	// pp.Println(kubecostResponse)
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

	url := Rancher_API + "/v3/projects"

	var bearer = "Bearer " + RancherBearerToken

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", bearer)

	// client := &http.Client{Transport: tr}
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

	// pp.Println("getting rancher projects from /v3/projects", RancherResponse)

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

// func bindnamespacestoprojects() (int, error) {
// 	var projectsNamespaces map[string][]string
// 	projectsNamespaces = make(map[string][]string)

// 	_ = projectsNamespaces

// 	Rancher_API, RancherBearerToken := getRancherAPIEnvVar()

// 	var bearer = "Bearer " + RancherBearerToken

// 	url := Rancher_API + "/v1/namespaces"

// 	req, err := http.NewRequest("GET", url, nil)
// 	req.Header.Add("Authorization", bearer)
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Println("No response from Rancher!", err)
// 	}
// 	defer resp.Body.Close()
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Println("error reading reponse from Rancher API when fetching namespaces from Rancher's API ", err)
// 	}
// 	RancherResponse := namespaceData{}
// 	jsonErr := json.Unmarshal(body, &RancherResponse)
// 	if jsonErr != nil {
// 		log.Fatal(jsonErr)
// 	}

// 	pp.Print("namespaceData", RancherResponse)

// 	// for _, namespaceData := range RancherResponse.Data {
// 	// 	fmt.Println("namespaceData.ID", namespaceData.ID)
// 	// 	fmt.Println("projectId", namespaceData.Metadata.Annotations.ProjectId)
// 	// 	projectsNamespaces[namespaceData.Metadata.Annotations.ProjectId] = []string{}
// 	// 	if (projectsNamespaces[namespaceData.Metadata.Annotations.ProjectId]) == nil {
// 	// 		var namespacelist []string
// 	// 		namespacelist = append(namespacelist, namespaceData.ID)
// 	// 		projectsNamespaces[namespaceData.Metadata.Annotations.ProjectId] = namespacelist

// 	// 	} else {
// 	// 		projectsNamespaces[namespaceData.Metadata.Annotations.ProjectId] = append(projectsNamespaces[namespaceData.Metadata.Annotations.ProjectId], namespaceData.ID)
// 	// 	}
// 	// }

// 	// fmt.Println("projectsNamespaces", projectsNamespaces)
// 	return 0, nil
// }

func generatebill() (BillingAccount, error) {
	// TODO: get this objects to be used in writing tests
	adminDetailsexample := []AdminDetails{}
	adminDetailsexample = append(adminDetailsexample, AdminDetails{Email: "exmaleadmin@email.com", Phone_number: "21452012", Name: "mohsen"})
	// adminDetailsexample["admin1"] = AdminDetails{email: "exmaleadmin@email.com", phone_number: "21452012", name: "mohsen"}

	// historyExample := make(map[time.Time]BillFile)
	// historyExample[time.Date(2022, time.June, 10, 9, 40, 0, 0, time.UTC)] = BillFile{pdfLink: "https://linktopdf.com/12540336", amount: 25}
	historyExample := make([]BillFile, 1)

	billFileExample := BillFile{
		BillingDate: time.Date(2022, time.June, 10, 9, 40, 0, 0, time.UTC),
		PdfLink:     "https://linktopdf.com/12540336",
		Amount:      25,
	}

	historyExample = append(historyExample, billFileExample)

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
		BillingAdmins: adminDetailsexample,
		// get the first date of any projects in the bill
		BillingStartDate: time.Now(),
		// company assigning project if exists
		Company:     Company{IsCompany: false, TaxId: "", Name: ""},
		AccountType: "Starter",
		// get balence from database
		Balance: 25.410,
		// lsit of previous bills
		History: historyExample,
		// is account suspended or not
		IsActive: true,
		// TODO: discuss if this value is better turned to map of clusters and projects whith clusters representing regions
		Projects: projectsexample,
	}

	fmt.Println("bill", bill)
	return bill, nil
}

func getK8SClient() *kubernetes.Clientset {
	// TODO: implement in cluster client
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	// kubeconfig = flag.String("kubeconfig", "", "./config/kubeconfig")
	// flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

func getnamespacesOfProjects(clientset *kubernetes.Clientset, ProjectId string) []string {

	var projectsNamespaces []string

	pp.Println("projectId", ProjectId)
	listOptions := metav1.ListOptions{
		// FieldSelector: "status.successful=1",
		LabelSelector: "field.cattle.io/projectId=" + ProjectId,
		// LabelSelector: "kubernetes.io/metadata.name=mehernamespace2",
		// LabelSelector: "field.cattle.io/projectId=p-9d4vx",
	}

	// pp.Println("listOptions", listOptions)

	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), listOptions)
	if err != nil {
		panic(err.Error())
	}

	// pp.Println("namespaces returned from kubernetes api", namespaces)
	for _, namespaceData := range namespaces.Items {
		projectsNamespaces = append(projectsNamespaces, namespaceData.ObjectMeta.Name)
	}

	pp.Println("projectsNamespaces", projectsNamespaces)

	return projectsNamespaces

}

func separateprojectIdfromClusterId(projectId string) string {
	if idx := strings.Index(projectId, ":"); idx != -1 {
		return projectId[idx+1:]
	}
	return projectId
}

// this functions updates the balence for the accounts on the pay-per-user plan
// func updateBalence() {}

type createBillingAccount struct {
	BillingAdmins []AdminDetails `json:"billingAdmins"`
	AccountType   AccountType    `json:"accountType"`
	Company       Company        `json:"company"`
	Projects      []Project      `json:"projects"`
}

func (h handler) CreateBillingAccount(c *gin.Context) {
	// log.Println("creating billing account by ", billingAccount.billingAdmins[0])

	// Validate input

	h.DB.Migrator().CreateTable(&BillingAccount{})

	var input createBillingAccount
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pp.Println("input", input)
	accountDetails := BillingAccount{
		UUID:             uuid.New(),
		BillingAdmins:    input.BillingAdmins,
		BillingStartDate: time.Now(),
		AccountType:      input.AccountType,
		Balance:          0.0,
		History:          []BillFile{},
		IsActive:         true,
		Company:          input.Company,
		Projects:         input.Projects,
	}

	if result := h.DB.Table("billing_accounts").Create(&accountDetails); result.Error != nil {
		c.AbortWithError(http.StatusNotFound, result.Error)
		return
	}

	pp.Println("accountDetails", accountDetails)
	c.JSON(http.StatusCreated, accountDetails)

}

// func getBillingAccount() {}

type UserRepo struct {
	Db *gorm.DB
}

// func New(sqldb *gorm.DB) *UserRepo {
// 	db := sqldb.InitDb()
// 	return &UserRepo{Db: db}
// }

type handler struct {
	DB *gorm.DB
}

// func exposeHttpServer(db *sql.DB) {
func exposeHttpServer(db *gorm.DB) {

	h := &handler{
		DB: db,
	}
	// 	BillingAdmins
	// History
	// Projects

	db.AutoMigrate(&BillingAccount{}, &AdminDetails{}, &BillFile{}, &Project{}, &Company{})

	// db.Migrator().CreateTable(&BillingAccount{})

	// db.Model(&BillingAccount).Related(&History)

	router := gin.Default()
	router.POST("/CreateBillingAccount", h.CreateBillingAccount)
	// router.GET("/getBillingAccount/:id", getBillingAccount)
	router.Run("localhost:8080")
}

func createDatabaseConnection() (*gorm.DB, error) {
	//TODO: take database info as env variables
	// connStr := "postgresql://bobsql:mypasswordbob@127.0.0.1/creometry-billing"
	// connStr := "postgresql://bobsql:mypasswordbob@127.0.0.1/creometry-billing?sslmode=disable"
	// Connect to database
	// db, err := sql.Open("postgres", connStr)
	// if err != nil {
	// 	log.Fatal(err)
	// 	return nil, err
	// }

	dsn := "host=localhost user=bobsql password=mypasswordbob dbname=creometry-billing port=5432 TimeZone=UTC+1"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return db, nil
}

// link namesapces to projects and then projects to users

// // we match projects to their respective users using https://tn.cloud.creometry.com/v3/projectRoleTemplateBindings?userId=${USERID}

// // from this one we'll link namespace to their respective projects using data[].metadata.annotations."field.cattle.io/projectId"
func main() {
	LoadDotEnvVariables()

	db, err := createDatabaseConnection()
	if err != nil {
		log.Fatal(err)
	}
	exposeHttpServer(db)

	// var namespaces map[string][]string
	// // var namespacesMetrics map[string][]string
	// namespaces = make(map[string][]string)
	// // RancherUsersDetails, err := getRancherUsers()
	// // if err != nil {
	// // 	log.Fatal(err)
	// // }
	// // matchUsersToProjects(RancherUsersDetails)

	// // bindnamespacestoprojects()
	// clientset := getK8SClient()

	// projects, error := getRancherProjects()
	// if error != nil {
	// 	log.Fatal(error)
	// }
	// for _, project := range projects.Data {
	// 	// getnamespacesOfProjects(clientset, project.ID)
	// 	projectId := separateprojectIdfromClusterId(project.ID)
	// 	namespaces[project.ID] = getnamespacesOfProjects(clientset, projectId)

	// 	for _, namespace := range namespaces[project.ID] {
	// 		metrics, err := getNamespaceMetrics(namespace)
	// 		if err != nil {
	// 			log.Printf("error while fetching metrics for namespace ", namespace, err)
	// 		}
	// 		// var testvariable kubecost.Allocation
	// 		// pp.Println("metrics", metrics)
	// 		pp.Println("metrics.Data[0]['__idle__']", metrics.Data[0]["__idle__"])
	// 		namespaceMetrics := Metrics{
	// 			CPUCoreHours:         metrics.Data[0]["__idle__"].CPUCoreHours,
	// 			CpuAverageUsage:      metrics.Data[0]["__idle__"].CPUCoreUsageAverage,
	// 			RamMinutes:           metrics.Data[0]["__idle__"].RAMByteHours,
	// 			RamAverageUsage:      metrics.Data[0]["__idle__"].RAMBytesUsageAverage,
	// 			networkTransferBytes: metrics.Data[0]["__idle__"].NetworkTransferBytes,
	// 			networkReceiveBytes:  metrics.Data[0]["__idle__"].NetworkTransferBytes,
	// 			//TODO: calculate PV total
	// 			// pvByteHours: metrics.Data[0]["__idle__"].PVs,
	// 		}
	// 		pp.Println("namespaceMetrics of ", namespace, namespaceMetrics)
	// 		// namespacesMetrics[namespace] =
	// 	}
	// }

	// generatebill()
	// clusterIds := getRancherClustersIds()

	// getRancherNamespaces(clusterIds)

	// // var namespaces = []string{"default"}
	// namespaceMetrics, err := getNamespaceMetrics()
	// if err != nil {
	// 	log.Println("error geeting namespace metrics from Kubecost's API", err)
	// }

	// fmt.Println("namespaceMetrics", namespaceMetrics)

}
