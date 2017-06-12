package main

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"strconv"
	"bytes"
	"fmt"
	"os"
	"github.com/gorilla/securecookie")

func mainHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.html") //setp 1
	t.Execute(w,"") //step 2
}

func getToken(request *http.Request) (t Token) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			t.NameIdNo, _ = strconv.Atoi(cookieValue["nameIdNo"])
			t.AccessToken = cookieValue["token"]
			t.UserName = cookieValue["name"]
		}
	}
	return
}

type FlexField struct{
	Name string `json:"attributeName"`
	Value interface{} `json:"value"`
}

type Object struct{
	FlexfieldCollection []FlexField `json:"flexfieldCollection"`
}

func (o Object) FlexMap () (result map[string]interface{}) {
	result = make(map[string]interface{})
	for _,v := range o.FlexfieldCollection{
		result[v.Name] = v.Value
	}
	return
}

type PolicyLine struct{
	PricePaid float64 `json:"pricePaid"`
	PolicyLineObjects []Object `json:"policyLineObjectCollection"`
}

type Policy struct{
	PolicyNo int `json:"policyNo"`
	PolicyStatus string `json:"policyStatus"`
	CancelCode int `json:"cancelCode"`
	YearStartDate string `json:"yearStartDate"`
	QuoteExpiryDate string `json:"quoteExpiryDate"`
	RenewalDate string `json:"renewalDate"`
        PolicyLineCollection []PolicyLine `json:"policyLineCollection"`
}

type PolicyList struct{
	PolicyCollection []Policy `json:"policyCollection"`
}


type ClearPolicy struct{
	InsuranceType string
	Period string
	Status string
	RegNo string
}

type ClearPolicyList []ClearPolicy

func GetPolicyList(nameIdNo int,token string)(cpl ClearPolicyList){
	url := fmt.Sprintf("http://195.230.98.205:8001/tiaws-rs/api/policies/?policyHolderId=%s&prodId=MT&productLineId=MT&objectTypeId=MT&coverDate=2017-06-13&allVersionYN=N", strconv.Itoa(nameIdNo))
	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	authToken := "Bearer "+token
	req.Header.Set("Authorization",authToken)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return ClearPolicyList{}
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return ClearPolicyList{}
	}
	defer resp.Body.Close()

	var policyList PolicyList
	if err := json.NewDecoder(resp.Body).Decode(&policyList); err != nil {
		log.Println(err)
	}

	for i,v := range policyList.PolicyCollection {
		if v.PolicyStatus == "P"{
			cp := ClearPolicy{}
			cp.InsuranceType = "MOTOR"
			cp.Period = v.YearStartDate + " - "+v.RenewalDate
			cp.Status = "Active"
			cp.RegNo = policyList.PolicyCollection[i].PolicyLineCollection[0].PolicyLineObjects[0].FlexMap()["REG_NO"].(string)
			cpl = append(cpl,cp)
		}
	}
	return
}



func policyHandler(w http.ResponseWriter, r *http.Request) {
	token := getToken(r)
	policies := GetPolicyList(token.NameIdNo,token.AccessToken)
	t, _ := template.ParseFiles("policy_list.html") //setp 1
	fmt.Println(len(policies))
	t.Execute(w,policies) //step 2
}
func newPolicyHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("policy_new.html") //setp 1
	t.Execute(w,"") //step 2
}

type RiskToSave struct{
	RiskNo int `json:"riskNo"`
	RiskYn string `json:"riskYn"`
	RiskExcess int `json:"riskExcess"`
	RiskSum int `json:"riskSum"`
	RiskFlex1 int `json:"riskFlex1,omitempty"`
}

type VarcharAttributeToSave struct{
	AttributeName string `json:"attributeName"`
	Value string `json:"value"`
}

type NumberAttributeToSave struct{
	AttributeName string `json:"attributeName"`
	Value int `json:"value"`
}

type DateAttributeToSave struct{
	AttributeName string `json:"attributeName"`
	Value string `json:"value"`
}

type FlexToSave struct{
	VarcharAttribute *VarcharAttributeToSave `json:"varcharAttribute,omitempty"`
	DateAttribute *DateAttributeToSave `json:"dateAttribute,omitempty"`
	NumberAttribute *NumberAttributeToSave `json:"numberAttribute,omitempty"`

}

type ObjectToSave struct{
	ObjectType string `json:"objectType"`
	TypeId string `json:"typeId"`
	ShortDesc string `json:"shortDesc"`
 	RiskCollection []RiskToSave `json:"riskCollection"`
	FlexfieldCollection []FlexToSave `json:"flexfieldCollection"`
}

type PolicyLineToSave struct{
	ProductLineId string `json:"productLineId"`
	ProductLineVerNo int `json:"productLineVerNo"`
	PolicyLineObjectCollection []ObjectToSave `json:"policyLineObjectCollection"`
}

type PolicyToSave struct{
	PolicyHolderId int `json:"policyHolderId"`
        ProdId string `json:"prodId"`
        PolicyStatus string `json:"policyStatus"`
        PaymentFrequency int `json:"paymentFrequency"`
        CoverStartDate string `json:"coverStartDate"`
        PolicyLineCollection  []PolicyLineToSave `json:"policyLineCollection"`
}

func NewPolicy(nameIdNo int, startDate string, carType string, RegNo string, Vin string)(p PolicyToSave){
	p.PolicyHolderId = nameIdNo
	p.ProdId = "MT"
	p.PolicyStatus = "P"
	p.PaymentFrequency = 12
	p.CoverStartDate = startDate
	pl := PolicyLineToSave{}
	pl.ProductLineId = "MT"
	pl.ProductLineVerNo = 1
	ob := ObjectToSave{}
	ob.ObjectType = "MT"
	ob.TypeId ="MT"
	ob.ShortDesc = "MIA Car Policy"
	r := RiskToSave{}
	r.RiskNo = 2
	r.RiskYn = "Y"
	r.RiskExcess = 50
	r.RiskSum = 500
	ob.RiskCollection = append(ob.RiskCollection, r)
	r = RiskToSave{}
	r.RiskNo = 3
	r.RiskYn = "Y"
	r.RiskExcess = 60
	r.RiskSum = 600
	ob.RiskCollection = append(ob.RiskCollection, r)
	r = RiskToSave{}
	r.RiskNo = 5
	r.RiskYn = "Y"
	r.RiskExcess = 50
	r.RiskSum = 500
	ob.RiskCollection = append(ob.RiskCollection, r)
	r = RiskToSave{}
	r.RiskNo = 7
	r.RiskYn = "Y"
	r.RiskExcess = 70
	r.RiskSum = 700
	r.RiskFlex1 = 100
	ob.RiskCollection = append(ob.RiskCollection, r)
	fl := FlexToSave{}
	ct := VarcharAttributeToSave{AttributeName:"CAR_TYPE",Value:carType}
	fl.VarcharAttribute = &ct
	ob.FlexfieldCollection = append(ob.FlexfieldCollection,fl)
	fl = FlexToSave{}
	bn := VarcharAttributeToSave{AttributeName:"BONUS",Value: "1"}
	fl.VarcharAttribute = &bn
	ob.FlexfieldCollection = append(ob.FlexfieldCollection,fl)
	fl = FlexToSave{}
	rn := VarcharAttributeToSave{AttributeName:"REG_NO",Value: RegNo}
	fl.VarcharAttribute = &rn
	ob.FlexfieldCollection = append(ob.FlexfieldCollection,fl)
	fl = FlexToSave{}
	bd := VarcharAttributeToSave{AttributeName:"BODY_NO",Value: Vin}
	fl.VarcharAttribute = &bd
	ob.FlexfieldCollection = append(ob.FlexfieldCollection,fl)
	fl = FlexToSave{}
	ag := VarcharAttributeToSave{AttributeName:"AGE_GROUP",Value : "1"}
	fl.VarcharAttribute = &ag
	ob.FlexfieldCollection = append(ob.FlexfieldCollection,fl)
	fl = FlexToSave{}
	vu := DateAttributeToSave{AttributeName:"VALID_UNTIL",Value: "2015-11-17"}
	fl.DateAttribute = &vu
	ob.FlexfieldCollection = append(ob.FlexfieldCollection,fl)
	fl = FlexToSave{}
	rd := DateAttributeToSave{AttributeName:"REG_DATE",Value : "2015-11-03"}
	fl.DateAttribute = &rd
	ob.FlexfieldCollection = append(ob.FlexfieldCollection,fl)
	fl = FlexToSave{}
	ee := NumberAttributeToSave{AttributeName :"EXTRA_EQUIPMENT_SUM",Value : 1000}
	fl.NumberAttribute = &ee
	ob.FlexfieldCollection = append(ob.FlexfieldCollection,fl)
	pl.PolicyLineObjectCollection = append(pl.PolicyLineObjectCollection,ob)
	p.PolicyLineCollection = append(p.PolicyLineCollection,pl)
	return
}

type CalcResult struct{
	Policy struct{
		PolicyNo int `json:"policyNo"`
		QuoteExpiryDate string `json:"quoteExpiryDate"`
		PaymentFrequency int `json:"paymentFrequency"`
		PolicyLineCollection []struct{
			PricePaid float64 `json:"pricePaid"`
		} `json:"policyLineCollection"`
	       } `json:"policy"`
	Result struct{
		ResultCode int `json:"resultCode"`
		MessageCollection []interface{} `json:"messageCollection"`
		CallDuration int `json:"callDuration"`
	       } `json:"result"`
}

func calculatePolicy(carType string,licensePlate string, vin string, startDate string, token string, nameIdNo int)(premium float64, queryExpiryDate string){
	url := "http://195.230.98.205:8001/tiaws-rs/api/policies/premium/"
	// Build the request
	policy, _ := json.Marshal(NewPolicy(nameIdNo,startDate,carType,licensePlate,vin))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(policy))
	authToken := "Bearer "+token
	req.Header.Set("Authorization",authToken)
	req.Header.Set("Content-Type","application/json")
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return 0,""
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return 0,""
	}
	defer resp.Body.Close()

	var calcResult CalcResult
	if err := json.NewDecoder(resp.Body).Decode(&calcResult); err != nil {
		log.Println(err)
	}

	premium = calcResult.Policy.PolicyLineCollection[0].PricePaid
	queryExpiryDate = calcResult.Policy.QuoteExpiryDate
	return
}

func calculatePolicyHandler(w http.ResponseWriter, r *http.Request) {
	token := getToken(r)
	premium, quoteExpiry := calculatePolicy(r.FormValue("car_type"),r.FormValue("license_plate"),r.FormValue("vin"),r.FormValue("start_date"), token.AccessToken, token.NameIdNo)

	t, _ := template.ParseFiles("quotation.html") //setp 1
	t.Execute(w,struct{
		Premium float64
		QuoteExpiry string
		CarType string
		LicensePlate string
		Vin string
		StartDate string}{
		Premium:premium,
		QuoteExpiry:quoteExpiry,
		CarType:r.FormValue("car_type"),
		LicensePlate:r.FormValue("license_plate"),
		Vin:r.FormValue("vin"),
		StartDate:r.FormValue("start_date"),
	}) //step 2
}

func CreatePolicy(carType string,licensePlate string, vin string, startDate string, token string, nameIdNo int)(err error){
	url := "http://195.230.98.205:8001/tiaws-rs/api/policies/"
	// Build the request
	policy, _ := json.Marshal(NewPolicy(nameIdNo,startDate,carType,licensePlate,vin))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(policy))
	authToken := "Bearer "+token
	req.Header.Set("Authorization",authToken)
	req.Header.Set("Content-Type","application/json")
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return err
	}
	defer resp.Body.Close()

	var calcResult CalcResult
	if err := json.NewDecoder(resp.Body).Decode(&calcResult); err != nil {
		log.Println(err)
	}
	return nil
}

func confirmPolicyHandler(w http.ResponseWriter, r *http.Request) {
	token := getToken(r)
	_ = CreatePolicy(r.FormValue("car_type"),r.FormValue("license_plate"),r.FormValue("vin"),r.FormValue("start_date"), token.AccessToken, token.NameIdNo)
	http.Redirect(w, r, "/policy/", http.StatusFound)
}

type Token struct{
	AccessToken string `json:"access_token"`
	NameIdNo int `json:nameIdNo"`
	UserName string
}

func createToken(userName string)(Token){
	url := fmt.Sprintf("http://195.230.98.205:8001/tiaws-rs/oauth/token?grant_type=password&client_id=MIA&client_secret=MIA&username=%s@affecto.com&password=tcc2016", userName)
	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return Token{}
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return Token{}
	}
	defer resp.Body.Close()

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		log.Println(err)
	}
	token.UserName = userName;
	log.Println("Token",token.AccessToken)
	log.Println("nameIdNo",token.NameIdNo)
	log.Println("UserName",token.UserName)
	return token
}

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func setSession(userName string, token string, nameIdNo int, response http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
		"token":token,
		"nameIdNo":strconv.Itoa(nameIdNo),
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	}
}

func loginHandler(response http.ResponseWriter, request *http.Request) {
	name := request.FormValue("email")
	pass := request.FormValue("password")
	redirectTarget := "/"
	log.Println(name,pass,redirectTarget)
	if name != "" && pass != "" && pass=="id17"{
		token := createToken(name)
		// .. check credentials ..
		setSession(token.UserName, token.AccessToken, token.NameIdNo, response)
		redirectTarget = "/policy/"
	}
	http.Redirect(response, request, redirectTarget, 302)
}

 func main() {
	 r := mux.NewRouter()
	 r.HandleFunc("/login/", loginHandler).Methods("POST")
	 r.HandleFunc("/", mainHandler).Methods("GET")
	 r.HandleFunc("/policy/", policyHandler).Methods("GET")
	 r.HandleFunc("/policy/new/", newPolicyHandler).Methods("GET")
	 r.HandleFunc("/policy/new/", calculatePolicyHandler).Methods("POST")
	 r.HandleFunc("/policy/new/confirm/", confirmPolicyHandler).Methods("POST")
	 r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))
	 http.Handle("/", r)

	 fmt.Println("Listening on port "+os.Getenv("HTTP_PLATFORM_PORT")+"....")
	 http.ListenAndServe(":"+os.Getenv("HTTP_PLATFORM_PORT"), nil)
 }
