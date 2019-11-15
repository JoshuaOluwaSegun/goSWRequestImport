package main

import (
	"sync"
	"time"

	apiLib "github.com/hornbill/goApiLib"
	"github.com/hornbill/sqlx"
)

const (
	version           = "1.7.1"
	appServiceManager = "com.hornbill.servicemanager"
)

var (
	appDBDriver            string
	cacheDBDriver          string
	arrCallsLogged         = make(map[string]string)
	arrCallDetailsMaps     = make([]map[string]interface{}, 0)
	boolConfLoaded         bool
	bufferMutex            = &sync.Mutex{}
	configFileName         string
	configDryRun           bool
	configDebug            bool
	configCustomerOrg      bool
	configMaxRoutines      string
	configVersion          bool
	connStrSysDB           string
	connStrAppDB           string
	espXmlmc               *apiLib.XmlmcInstStruct
	counters               counterTypeStruct
	mapGenericConf         swCallConfStruct
	analysts               []analystListStruct
	categories             []categoryListStruct
	closeCategories        []categoryListStruct
	customers              []customerListStruct
	companies              []companyListStruct
	priorities             []priorityListStruct
	services               []serviceListStruct
	sites                  []siteListStruct
	teams                  []teamListStruct
	sqlCallQuery           string
	swImportConf           swImportConfStruct
	timeNow                string
	startTime              time.Time
	endTime                time.Duration
	xmlmcInstanceConfig    xmlmcConfigStruct
	mutex                  = &sync.Mutex{}
	mutexAnalysts          = &sync.Mutex{}
	mutexArrCallsLogged    = &sync.Mutex{}
	mutexBar               = &sync.Mutex{}
	mutexCategories        = &sync.Mutex{}
	mutexCloseCategories   = &sync.Mutex{}
	mutexCounters          = &sync.Mutex{}
	mutexCustomers         = &sync.Mutex{}
	mutexCompanies         = &sync.Mutex{}
	mutexPriorities        = &sync.Mutex{}
	mutexServices          = &sync.Mutex{}
	mutexSites             = &sync.Mutex{}
	mutexTeams             = &sync.Mutex{}
	reqPrefix              string
	maxGoroutines          = 1
	boolProcessAttachments bool
	dbapp                  *sqlx.DB
	dbsys                  *sqlx.DB
)

// ----- Structures -----
type counterTypeStruct struct {
	sync.Mutex
	created        int
	createdSkipped int
	filesAttached  int
}

//----- Config Data Structs
type swImportConfStruct struct {
	HBConf                    hbConfStruct //Hornbill Instance connection details
	SWServerAddress           string
	AttachmentRoot            string
	CustomerType              string
	SMProfileCodeSeperator    string
	RelatedRequestQuery       string
	SWSystemDBConf            sysDBConfStruct //Cache Data (sw_systemdb) connection details
	SWAppDBConf               appDBConfStruct //App Data (swdata) connection details
	RequestTypesToImport      []swCallConfStruct
	PriorityMapping           map[string]interface{}
	TeamMapping               map[string]interface{}
	CategoryMapping           map[string]interface{}
	ResolutionCategoryMapping map[string]interface{}
	ServiceMapping            map[string]interface{}
	StatusMapping             map[string]interface{}
}
type hbConfStruct struct {
	InstanceID string
	UserName   string
	Password   string
}
type refStruct struct {
	MasterRef string
	SlaveRef  string
}
type sysDBConfStruct struct {
	Driver   string
	UserName string
	Password string
}
type appDBConfStruct struct {
	Driver   string
	Server   string
	UserName string
	Password string
	Port     int
	Database string
	Encrypt  bool
}
type swCallConfStruct struct {
	Import                 bool
	CallClass              string
	SupportworksCallClass  string
	DefaultTeam            string
	DefaultPriority        string
	DefaultService         string
	SQLStatement           string
	CoreFieldMapping       map[string]interface{}
	AdditionalFieldMapping map[string]interface{}
}

//----- XMLMC Config and Interaction Structs
type xmlmcConfigStruct struct {
	instance string
	url      string
	zone     string
}
type xmlmcResponse struct {
	MethodResult string      `xml:"status,attr"`
	State        stateStruct `xml:"state"`
}

//----- Shared Structs -----
type stateStruct struct {
	Code     string `xml:"code"`
	ErrorRet string `xml:"error"`
}

//----- Data Structs -----

type xmlmcSysSettingResponse struct {
	MethodResult string      `xml:"status,attr"`
	State        stateStruct `xml:"state"`
	Setting      string      `xml:"params>option>value"`
}

//----- Request Logged Structs
type xmlmcRequestResponseStruct struct {
	MethodResult string      `xml:"status,attr"`
	RequestID    string      `xml:"params>primaryEntityData>record>h_pk_reference"`
	SiteCountry  string      `xml:"params>rowData>row>h_country"`
	Diags        []string    `xml:"diagnostic>log"`
	State        stateStruct `xml:"state"`
}
type xmlmcBPMSpawnedStruct struct {
	MethodResult string      `xml:"status,attr"`
	Identifier   string      `xml:"params>identifier"`
	State        stateStruct `xml:"state"`
}

//----- Site Structs
type siteListStruct struct {
	SiteName string
	SiteID   int
}
type xmlmcSiteListResponse struct {
	MethodResult string      `xml:"status,attr"`
	SiteID       int         `xml:"params>rowData>row>h_id"`
	SiteName     string      `xml:"params>rowData>row>h_site_name"`
	SiteCountry  string      `xml:"params>rowData>row>h_country"`
	State        stateStruct `xml:"state"`
}

//----- Priority Structs
type priorityListStruct struct {
	PriorityName string
	PriorityID   int
}
type xmlmcPriorityListResponse struct {
	MethodResult string      `xml:"status,attr"`
	PriorityID   int         `xml:"params>rowData>row>h_pk_priorityid"`
	PriorityName string      `xml:"params>rowData>row>h_priorityname"`
	State        stateStruct `xml:"state"`
}

//----- Service Structs
type serviceListStruct struct {
	ServiceName          string
	ServiceID            int
	ServiceBPMIncident   string
	ServiceBPMService    string
	ServiceBPMChange     string
	ServiceBPMProblem    string
	ServiceBPMKnownError string
}
type xmlmcServiceListResponse struct {
	MethodResult  string      `xml:"status,attr"`
	ServiceID     int         `xml:"params>rowData>row>h_pk_serviceid"`
	ServiceName   string      `xml:"params>rowData>row>h_servicename"`
	BPMIncident   string      `xml:"params>rowData>row>h_incident_bpm_name"`
	BPMService    string      `xml:"params>rowData>row>h_service_bpm_name"`
	BPMChange     string      `xml:"params>rowData>row>h_change_bpm_name"`
	BPMProblem    string      `xml:"params>rowData>row>h_problem_bpm_name"`
	BPMKnownError string      `xml:"params>rowData>row>h_knownerror_bpm_name"`
	State         stateStruct `xml:"state"`
}

//----- Team Structs
type teamListStruct struct {
	TeamName string
	TeamID   string
}
type xmlmcTeamListResponse struct {
	MethodResult string      `xml:"status,attr"`
	TeamID       string      `xml:"params>rowData>row>h_id"`
	TeamName     string      `xml:"params>rowData>row>h_name"`
	State        stateStruct `xml:"state"`
}

//----- Category Structs
type categoryListStruct struct {
	CategoryCode string
	CategoryID   string
	CategoryName string
}
type xmlmcCategoryListResponse struct {
	MethodResult string      `xml:"status,attr"`
	CategoryID   string      `xml:"params>id"`
	CategoryName string      `xml:"params>fullname"`
	State        stateStruct `xml:"state"`
}

//----- Analyst Structs
type analystListStruct struct {
	AnalystID   string
	AnalystName string
}
type xmlmcAnalystListResponse struct {
	MethodResult     string      `xml:"status,attr"`
	AnalystFullName  string      `xml:"params>name"`
	AnalystFirstName string      `xml:"params>firstName"`
	AnalystLastName  string      `xml:"params>lastName"`
	State            stateStruct `xml:"state"`
}

//----- Customer Structs
type customerListStruct struct {
	CustomerHornbillID string
	CustomerID         string
	CustomerName       string
	CustomerOrgID      string
}
type companyListStruct struct {
	OrgID       string
	ContainerID string
}
type xmlmcOrgListResponse struct {
	MethodResult string                 `xml:"status,attr"`
	RowResult    []xmlCompanyListStruct `xml:"params>rowData>row"`
	State        stateStruct            `xml:"state"`
}
type xmlCompanyListStruct struct {
	OrgID       string `xml:"h_organization_id"`
	ContainerID string `xml:"h_id"`
}

type xmlmcContactListResponse struct {
	MethodResult       string      `xml:"status,attr"`
	CustomerFirstName  string      `xml:"params>rowData>row>h_firstname"`
	CustomerLastName   string      `xml:"params>rowData>row>h_lastname"`
	CustomerOrgID      string      `xml:"params>rowData>row>h_organization_id"`
	CustomerHornbillID string      `xml:"params>rowData>row>h_pk_id"`
	State              stateStruct `xml:"state"`
}
type xmlmcCustomerListResponse struct {
	MethodResult       string      `xml:"status,attr"`
	CustomerFirstName  string      `xml:"params>rowData>row>h_first_name"`
	CustomerLastName   string      `xml:"params>rowData>row>h_last_name"`
	CustomerOrgID      string      `xml:"params>rowData>row>h_organization_id"` // probably not used at all
	CustomerHornbillID string      `xml:"params>rowData>row>h_user_id"`
	State              stateStruct `xml:"state"`
}

//----- Associated Record Struct
type reqRelStruct struct {
	MasterRef string `db:"parentRequest"`
	SlaveRef  string `db:"childRequest"`
}

//----- File Attachment Structs
type xmlmcAttachmentResponse struct {
	MethodResult    string      `xml:"status,attr"`
	ContentLocation string      `xml:"params>contentLocation"`
	State           stateStruct `xml:"state"`
	HistFileID      string      `xml:"params>primaryEntityData>record>h_pk_fileid"`
}

//----- Email Attachment Structs
type xmlmcEmailAttachmentResponse struct {
	MethodResult    string             `xml:"status,attr"`
	Recipients      []recipientStruct  `xml:"params>recipient"`
	RFCHeader       string             `xml:"params>rfcHeader"`
	Subject         string             `xml:"params>subject"`
	Body            string             `xml:"params>body"`
	HTMLBody        string             `xml:"params>htmlBody"`
	TimeSent        string             `xml:"params>timeSent"`
	FileAttachments []fileAttachStruct `xml:"params>fileAttachment"`
	State           stateStruct        `xml:"state"`
}

type fileAttachStruct struct {
	FileName  string `xml:"fileName"`
	FileData  string `xml:"fileData"`
	FileSize  string `xml:"fileSize"`
	TimeStamp string `xml:"timeStamp"`
	MIMEType  string `xml:"contentType"`
	ContentID string `xml:"contentId"`
}
type recipientStruct struct {
	Class   string `xml:"class"`
	Address string `xml:"address"`
	Name    string `xml:"name"`
}

type swmStruct struct {
	Content     string
	Subject     string
	Attachments []fileAttachStruct
}

//----- File Attachment Struct
type fileAssocStruct struct {
	ImportRef       int
	SmCallRef       string
	FileID          string  `db:"fileid"`
	CallRef         string  `db:"callref"`
	DataID          string  `db:"dataid"`
	UpdateID        string  `db:"updateid"`
	Compressed      string  `db:"compressed"`
	SizeU           float64 `db:"sizeu"`
	SizeC           float64 `db:"sizec"`
	FileName        string  `db:"filename"`
	AddedBy         string  `db:"addedby"`
	TimeAdded       string  `db:"timeadded"`
	FileTime        string  `db:"filetime"`
	FileData        string
	Extension       string
	Description     string
	EmailAttachment fileAttachStruct
}

type xmlmcIndexListResponse struct {
	MethodResult string      `xml:"status,attr"`
	State        stateStruct `xml:"state"`
	Indexes      []string    `xml:"params>indexStorage"`
}

//RequestDetails struct for chan
type RequestDetails struct {
	CallClass string
	CallMap   map[string]interface{}
	SwCallID  string
}

//RequestReferences struct for chan
type RequestReferences struct {
	SmCallID string
	SwCallID string
}
