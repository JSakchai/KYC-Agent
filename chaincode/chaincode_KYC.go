/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at
  http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/



package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/derekparker/delve/pkg/dwarf/line"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var customerIndexStr = "_customerindex" //name for the key/value that will store a list of all known customers
var brokerIndexStr = "_brokerindex"     //name for the key/value that will store a list of all known customers

// key of customer
var customerKey = "cus_"

// BrokerKey key of broker
var BrokerKey = "bro_"

// GuaranteeIDKey key of guarantee id
var GuaranteeIDKey = "gua_"

// CusGuaIDKey key of something
var CusGuaIDKey = "cgi_"

//Customer is Customer
type Customer struct {
	Name       string `json:"name"` //the fieldtags are needed to keep case from bouncing around
	CardID     string `json:"cardid"`
	TelNo      string `json:"telno"`
	Age        int    `json:"age"`
	Occupation string `json:"occupation"`
	Address    string `json:"address"`
	Creator    string `json:"creator"`
}

//GuaranteeID generate from Customer
type GuaranteeID struct {
	GuaranteeID  string `json:"guaranteeid"`
	CustomerID   string `json:"customerid"`
	AllowBroke   []int  `json:"allowbroke"`
	PendingBroke []int  `json:"pendingbroke"`
}

// type CusGuaID struct {
// 	CardID      string      `json:"cardid"`
// 	Customer    Customer    `json:"customer"`
// 	GuaranteeID GuaranteeID `json:"guaranteeid"`
// }

// Broker Contain Name and Number and AllowCustomer
type Broker struct {
	Name            string   `json:"name"`
	BrokerNo        int      `json:"brokerno"`
	AllowCustomer   []string `json:"allowcustomer"`
	PendingCustomer []string `json:"pendingcustomer"`
	rejectCustomer	[]string `json:"rejectcustomer"`
}

// Main
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init - reset all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var Aval int
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("kyc", []byte(strconv.Itoa(Aval))) //making a test var "kyc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}

	var empty []string
	jsonAsBytes, _ := json.Marshal(empty) //marshal an emtpy array of strings to clear the index
	err = stub.PutState(customerIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	err = stub.PutState(brokerIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Run - Our entry point for Invocations - [LEGACY] obc-peer 4/25/2016
// ============================================================================================================================
func (t *SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "write" { //writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "newcustomer" { //create a new customer
		return t.newcustomer(stub, args)
	} else if function == "newbroke" {
		return t.newbroke(stub, args)
	} else if function == "requestPermission" {
		return t.requestPermission(stub, args)
	} else if function == "customerallow" {
		return t.customerallow(stub, args)
	} else if function == "cancelAllow"{
		return  t.cancelAllow(stub,args)
	} else if function == "rejectBroker" {
		return t.rejectBroker(stub, args)
	}
	fmt.Println("invoke did not find func: " + function) //error

	return nil, errors.New("Received unknown function invocation")
}

// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	}
	if function == "readcustomer" { //read a variable
		return t.readcustomer(stub, args)
	}
	if function == "readcustomergid" {
		return t.readcustomergid(stub, args)
	}
	if function == "readbroker" { //read a variable
		return t.readbroker(stub, args)
	}
	fmt.Println("query did not find func: " + function) //error

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name) //get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil //send it onward
}

// ============================================================================================================================
// Read - read a variable from chaincode state by cardid
// ============================================================================================================================
func (t *SimpleChaincode) readcustomer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var cardid, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	cardid = args[0]
	valAsbytes, err := stub.GetState(customerKey + cardid) //get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + cardid + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil //send it onward
}

// ============================================================================================================================
// Read - read a variable from chaincode state by guaranteeid
// ============================================================================================================================
func (t *SimpleChaincode) readcustomergid(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var gid, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	gid = args[0]
	valAsbytes, err := stub.GetState(GuaranteeIDKey + gid) //get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + gid + "\"}"
		return nil, errors.New(jsonResp)
	}

	gua := GuaranteeID{}
	json.Unmarshal(valAsbytes, &gua)

	return valAsbytes, nil //send it onward
}

// ============================================================================================================================
// Read - read a variable from chaincode state by brokeno
// ============================================================================================================================
func (t *SimpleChaincode) readbroker(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var brokeno, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	brokeno = args[0]
	valAsbytes, err := stub.GetState(BrokerKey + brokeno) //get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + brokeno + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil //send it onward
}

func (t *SimpleChaincode) requestPermission(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	// var brokerno, gid string
	// var err error

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting guarantee id and brokeno")
	}

	gid := args[0]
	//brokeno := args[1]
	brokeNoAsString := args[1]
	brokeNo, err := strconv.Atoi(brokeNoAsString)
	brokerAsBytes, err := stub.GetState(BrokerKey + brokeNoAsString)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for broker " + brokeNoAsString + "\"}"
		return nil, errors.New(jsonResp)
	}
	broker := Broker{}
	json.Unmarshal(brokerAsBytes, &broker)

	gidAsbytes, err := stub.GetState(GuaranteeIDKey + gid)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + gid + "\"}"
		return nil, errors.New(jsonResp)
	}
	gua := GuaranteeID{}
	json.Unmarshal(gidAsbytes, &gua)

	already := false
	fmt.Printf("GID %s\n", gidAsbytes)
	fmt.Printf("brokeno %d\n", brokeNo)
	for _, s := range gua.AllowBroke {
		//set[s] = struct{}{}
		if s == brokeNo {
			already = true
		}
	}

	fmt.Println("already" + strconv.FormatBool(already))

	if already {
		jsonResp := "{\"Error\":\"Already Allowed " + brokeNoAsString + "\"}"
		return nil, errors.New(jsonResp)
	}

	gua.PendingBroke = append(gua.PendingBroke, brokeNo)
	broker.PendingCustomer = append(broker.PendingCustomer, gua.GuaranteeID)

	fmt.Println("gua.AllowBroke")
	fmt.Println(gua.AllowBroke)

	jsonAsBytes, err := json.Marshal(gua)
	fmt.Printf("GID %s\n", jsonAsBytes)
	err = stub.PutState(GuaranteeIDKey+gid, jsonAsBytes) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}

	jsonBrokerAsBytes, err := json.Marshal(broker)
	fmt.Printf("Broke %s\n", jsonBrokerAsBytes)
	err = stub.PutState(BrokerKey+brokeNoAsString, jsonBrokerAsBytes) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ============================================================================================================================
// Set User Permission on Customer
// ============================================================================================================================
func (t *SimpleChaincode) customerallow(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	//gid, brokeno

	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	gid := args[0]
	brokeNoAsString := args[1]
	brokeNo, err := strconv.Atoi(brokeNoAsString)

	gidAsBytes, err := stub.GetState(GuaranteeIDKey + gid)
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	gua := GuaranteeID{}
	json.Unmarshal(gidAsBytes, &gua)

	brokeAsBytes, err := stub.GetState(BrokerKey + brokeNoAsString)
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	broke := Broker{}
	json.Unmarshal(brokeAsBytes, &broke)

	for i := len(gua.PendingBroke) - 1; i >= 0; i-- {
		if gua.PendingBroke[i] == brokeNo {
			gua.PendingBroke = append(gua.PendingBroke[:i], gua.PendingBroke[i+1:]...)
			break
		}
	}
	for i := len(broke.PendingCustomer) - 1; i >= 0; i-- {
		if broke.PendingCustomer[i] == gid {
			broke.PendingCustomer = append(broke.PendingCustomer[:i], broke.PendingCustomer[i+1:]...)
			break
		}
	}
	already := false
	for _, s := range gua.AllowBroke {
		if s == brokeNo {
			already = true
		}
	}
	if !already {
		gua.AllowBroke = append(gua.AllowBroke, brokeNo)
		broke.AllowCustomer = append(broke.AllowCustomer, gid)
	}

	//write state

	jsonAsBytes, _ := json.Marshal(broke)
	err = stub.PutState(BrokerKey+brokeNoAsString, jsonAsBytes) //rewrite the customer with id as key
	if err != nil {
		return nil, err
	}

	jsonGuaAsBytes, _ := json.Marshal(gua)
	err = stub.PutState(GuaranteeIDKey+gid, jsonGuaAsBytes) //rewrite the customer with id as key
	if err != nil {
		return nil, err
	}

	fmt.Println("- end set allow permission")
	return nil, nil
}
// ==================================================================================================================
// cancel Allow broker
// ==================================================================================================================
func (t *SimpleChaincode) cancelAllow(stub shim.ChaincodeStubInterface, args []string) ([]byte , error){
	var err error
	if len(args) == 0 {
		return nil,errors.New("Argument is NULL")
	}else if len(args[0]) == 0 {
		return nil, errors.New("Argument is NULL")
	}else if len(args[1]) == 0 {
		return nil, errors.New("Argument is NULL")
	}
	gid := args[0]
	brkstring := args[1]
	brkNO,err :=  strconv.Atoi(brkstring)
	resGID := GuaranteeID{}
	customerASByte,err := stub.GetState(customerKey + gid)
	if err != nil {
		return nil, errors.New("Can't gate state")
	}
	json.Unmarshal(customerASByte,&resGID)
	for i := len(resGID.AllowBroke) -1 ;i < 0 ;i-- {
		if resGID.AllowBroke[i] == brkNO {
			delete(resGID.AllowBroke,brkNO) //remove broker number from allowbroker of  Guarantee structure
			append(resGID.PendingBroke,brkNO) //add broker number to pendingBroker of Guarantee structure
		}
	}
	brokerAsBye,err := stub.GetState(BrokerKey + brkNO)
	resBrk := Broker{}
	json.Unmarshal(brokerAsBye,&resBrk)
	for i := len(resBrk.AllowCustomer) -1;i<0;i-- {
		if resBrk.AllowCustomer[i] == gid {
			delete(resBrk.AllowCustomer[i],gid) //remove guarantee ID from allowCustomer of Broker structure
			append(resBrk.PendingCustomer,gid) //add guarantee ID form pendingCustomer of Broker struct

		}
	}
	// write Guarantee ID
	jsonGIDAsByte, _ := json.Marshal(resGID)
	err = stub.PutState(GuaranteeIDKey + gid,jsonGIDAsByte)
	if err != nil {
		return nil, errors.New("Can't put state Guarantee ID")
	}
	jsonBrkAsByte, _ := json.Marshal(resBrk)
	err = stub.PutState(BrokerKey + string(brkNO),jsonBrkAsByte)
	if err != nil {
		return nil, errors.New("Can't put state Broker Number")
	}
	return nil,nil
}
// =================================================================================================================
//Reject Broker nubmer
// =================================================================================================================
func (t *SimpleChaincode) rejectBroker(stub shim.ChaincodeStubInterface, args []string) ([]byte, error){
	var err error

	if len(args) == 0 {
		return nil, errors.New("Argument is NULL")
	}else if len(args[0]) == 0{
		return nil,errors.New("Argument is NULL")
	}else if len(args) == 0 {
		return nil, errors.New("Argument is NULL")
	}
	gid := args[0]
	brkstring := args[1]
	brkNO,err := strconv.Atoi(brkstring)
	GIDAsByte,err := stub.GetState(GuaranteeIDKey + gid)
	resGID := GuaranteeID{}
	json.Unmarshal(GIDAsByte,&resGID)
	for i := len(resGID.PendingBroke) -1 ;i < 0 ;i-- {
		if resGID.PendingBroke[i] == brkNO {
			delete(resGID.PendingBroke, gid)  //remove Broker from peddind of struct Guarantee
		}
	}
	brkAsByte,err := stub.GetState(BrokerKey + brkNO)
	resBrk := Broker{}
	json.Unmarshal(brkAsByte,&resBrk)
	for i := len(resBrk.PendingCustomer) -1 ;i <0;i-- {
		if resBrk.PendingCustomer[i] == gid {
			delete(resBrk.PendingCustomer,gid) //remove from pending of struct Broker
			append(resBrk.rejectCustomer,gid) //add to reject broker of struct broker
		}
	}
	//write key guarantee ID
	jsonGIDAsbyte,_ := json.Marshal(resGID)
	err = stub.PutState(GuaranteeIDKey + gid,jsonGIDAsbyte)
	if err != nil  {
		return nil, errors.New("Can't put state  Guarantee ID")
	}
	//write key broker number
	jsonBrkAsByte,_ := json.Marshal(resBrk)
	err = stub.PutState(BrokerKey + string(brkNO), jsonBrkAsByte)
	if err != nil {
		return nil, errors.New("Can't put stat Broker number")
	}

	return nil,nil
}
// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	name = args[0] //rename for funsies
	value = args[1]
	err = stub.PutState(name, []byte(value)) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// Init Customer - create a new customer, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) newcustomer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//   0        1       2        3           4          5
	// "name", "telno", "age", "occupation", "cardid", "creator"
	if len(args) != 6 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	//input sanitation
	fmt.Println("- start init customer")
	if len(args[0]) <= 0 {
		fmt.Println("- not pass0 init customer")
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		fmt.Println("- not pass1 init customer")
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		fmt.Println("- not pass2 init customer")
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		fmt.Println("- not pass3 init customer")
		return nil, errors.New("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		fmt.Println("- not pass4 init customer")
		return nil, errors.New("5th argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		fmt.Println("- not pass5 init customer")
		return nil, errors.New("6th argument must be a non-empty string")
	}

	fmt.Println("- pass1 init customer")
	name := args[0]
	telno := strings.ToLower(args[1])
	age, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New("3rd argument must be a numeric string")
	}
	occupation := strings.ToLower(args[3])
	cardid := args[4]
	creator := args[5]

	fmt.Println("- pass2 init customer")
	//check if customer already exists
	customerAsBytes, err := stub.GetState(customerKey + name)
	if err != nil {
		return nil, errors.New("Failed to get customer name")
	}
	res := Customer{}
	json.Unmarshal(customerAsBytes, &res)
	if res.Name == name {
		fmt.Println("This customer arleady exists: " + name)
		fmt.Println(res)
		return nil, errors.New("This customer arleady exists") //all stop a customer by this name exists
	}

	res.Name = name
	res.TelNo = telno
	res.Age = age
	res.Occupation = occupation
	res.CardID = cardid
	res.Creator = creator
	//build the customer json string manually
	//str := `{"name": "` + name + `", "telno": "` + telno + `", "size": ` + strconv.Itoa(size) + `, "user": "` + user + `"}`
	//err = stub.PutState(name, []byte(str)) //store customer with id as key
	str, err := json.Marshal(res)
	err = stub.PutState(customerKey+name, str)
	if err != nil {
		return nil, err
	}
	err = stub.PutState(customerKey+cardid, str)
	if err != nil {
		return nil, err
	}

	sha256AsByte := sha256.Sum256(str)

	guaranteeID := GuaranteeID{}
	var emptyIntArray []int
	guaranteeID.GuaranteeID = strings.ToUpper(hex.EncodeToString(sha256AsByte[:]))
	guaranteeID.CustomerID = res.CardID
	guaranteeID.AllowBroke = emptyIntArray
	guaranteeID.PendingBroke = emptyIntArray

	str, err = json.Marshal(guaranteeID)
	err = stub.PutState(GuaranteeIDKey+guaranteeID.GuaranteeID, str)
	if err != nil {
		return nil, err
	}

	//gid := GuaranteeID{}
	// h := sha1.New()
	// h.Write([]byte(res))
	// gid.GuaranteeID = h.Sum()

	//get the customer index-
	customersAsBytes, err := stub.GetState(customerIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get customer index")
	}
	var customerIndex []string
	json.Unmarshal(customersAsBytes, &customerIndex) //un stringify it aka JSON.parse()

	//append
	customerIndex = append(customerIndex, cardid) //add customer name to index list
	fmt.Println("! customer index: ", customerIndex)
	jsonAsBytes, _ := json.Marshal(customerIndex)
	err = stub.PutState(customerIndexStr, jsonAsBytes) //store name of customer

	fmt.Println("- end init customer")
	return nil, nil
}

// ============================================================================================================================
// Init Broke - create a new broke, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) newbroke(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//   0        1
	// "name", "brokeID"
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	//input sanitation
	fmt.Println("- start init customer")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}

	name := args[0]
	brokeNoAsString := args[1]
	brokeNo, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New("2nd argument must be a numeric string")
	}

	//check if broker already exists
	brokerAsBytes, err := stub.GetState(brokeNoAsString)
	if err != nil {
		return nil, errors.New("Failed to get broker name")
	}
	res := Broker{}
	json.Unmarshal(brokerAsBytes, &res)
	if res.BrokerNo == brokeNo {
		fmt.Println("This broker arleady exists: " + name)
		fmt.Println(res)
		return nil, errors.New("This broker arleady exists") //all stop a broker by this name exists
	}

	res.Name = name
	res.BrokerNo = brokeNo
	//build the broker json string manually
	//str := `{"name": "` + name + `", "telno": "` + telno + `", "size": ` + strconv.Itoa(size) + `, "user": "` + user + `"}`
	//err = stub.PutState(name, []byte(str)) //store broker with id as key
	str, err := json.Marshal(res)
	err = stub.PutState(BrokerKey+brokeNoAsString, str)
	if err != nil {
		return nil, err
	}

	//get the broker index-
	brokersAsBytes, err := stub.GetState(brokerIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get broker index")
	}
	var brokerIndex []string
	json.Unmarshal(brokersAsBytes, &brokerIndex) //un stringify it aka JSON.parse()

	//append
	brokerIndex = append(brokerIndex, brokeNoAsString) //add broker name to index list
	fmt.Println("! broker index: ", brokerIndex)
	jsonAsBytes, _ := json.Marshal(brokerIndex)
	err = stub.PutState(brokerIndexStr, jsonAsBytes) //store name of broker

	fmt.Println("- end init broker")
	return nil, nil
}