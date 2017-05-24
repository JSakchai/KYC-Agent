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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	//"github.com/drone/routes"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	//"encoding/gob"
	//"crypto/rand"
	//"github.com/fabric/core/ledger/statemgmt"
	//"github.com/fabric/core/chaincode"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var customerIndexStr = "_customerindex" //name for the key/value that will store a list of all known customers
var openTradesStr = "_opentrades"       //name for the key/value that will store all open trades

type Customer struct {
	CardId	    string   `json:"cardid"`
	Name        string   `json:"name"` //the fieldtags are needed to keep case from bouncing around
	TelNo       string   `json:"telno"`
	Age         int      `json:"age"`
	Birthday    string   `json:"birthday"`
	Occupation  string   `json:"occupation"`
	Address     string   `json:"address"`
	AllowBroke  []Broker `json:"allowbroke"`
	GauranteeID string   `json:"gauranteeid"`
}

type Broker struct {
	BrokerNo      string      `json:"brokerno"`
	BrokerName    string     `json:"brokername"`
	AllowCustomer []string `json:"allowcustomer"`
}
type  guaranteeID struct {
	GuaranteeID 	string `json:"guarantee_id"`
	CardID		string	`json:"cardid"`
	Name 		string	`json:"name"`
	ExDate		string	`json:"expire_date"`
	isActive	string	`json:"isActive"`
	AllowBroke	[]string `json:"allow_broke"`
	PendingBroke	[]string `json:"pending_broke"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
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

	/*var trades AllTrades
	jsonAsBytes, _ = json.Marshal(trades) //clear the open trade struct
	err = stub.PutState(openTradesStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	*/
	return nil, nil
}

// ============================================================================================================================
// Run - Our entry point for Invocations - [LEGACY] obc-peer 4/25/2016
// ============================================================================================================================
func (t *SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
		/*} else if function == "delete" { //deletes an entity from its state
		res, err := t.Delete(stub, args)
		cleanTrades(stub) //lets make sure all open trades are still valid
		return res, err*/
	} else if function == "write" { //writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "new_customer" { //create a new customer
		return t.new_customer(stub, args)
	} else if function == "set_user" { //change owner of a customer
		res, err := t.set_user(stub, args)
		//cleanTrades(stub) //lets make sure all open trades are still valid
		return res, err
	} else if function == "open_trade" { //create a new trade order
		//return t.open_trade(stub, args)
	} else if function == "perform_trade" { //forfill an open trade order
		// res, err := t.perform_trade(stub, args)
		// cleanTrades(stub) //lets clean just in case
		// return res, err
	} else if function == "remove_trade" { //cancel an open trade order
		// return t.remove_trade(stub, args)
	}else if function == "update_customer"{
		res,err := t.update_customer(stub,args)
		return  res ,err
	}else if function == "newBroker"{
		res, err := t.newBroker(stub,args)
		return res,err
	}else if function == "updateBroker"{
		res, err := t.updateBroker(stub,args)
		return  res, err
	}else if function == "updateAllowBroker"{
		res,err := t.updateAllowBroker(stub,args)
		return  res,err
	}
	fmt.Println("invoke did not find func: " + function) //error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
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
// ================================================================================================================
// Update customer
// =================================================================================================================
func (t *SimpleChaincode)  update_customer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error){
	var err error
	// check argument

	cardid := args[0]
	name := args[1]
	telno := strings.ToLower(args[2])
	age, err := strconv.Atoi(args[3])
	birthday := args[4]
	occupation := strings.ToLower(args[5])
	if err != nil {
		return nil, errors.New("3rd argument must be a numeric string")
	}
	address := args[6]
	valAsBytes,err := stub.GetState(cardid) //get date by key in json from block
	if err != nil {
		return  nil,errors.New("Fiail get name from json")
	}
	res := Customer{}
	json.Unmarshal([]byte(valAsBytes),&res)  //json to bytes and keep address to variable res
	if  res.CardId == cardid {
		if len(args[0]) <= 0{
			return  nil , errors.New("The card id parameter wrong")
		}else if len(args[1]) <= 0 {
			return  nil, errors.New("the Name Parameter wrong")
		}else if len(args[2]) <= 0{
			return  nil, errors.New("the Telno Parameter wrong")
		}else if len(args[3]) <= 0 {
			return  nil, errors.New("the Age Parameter wrong")
		}else if len(args[4]) <= 0 {
			return  nil, errors.New("the birthday Parameter wrong ")
		}else if len(args[5]) <= 0{
			return  nil, errors.New("the occupation Parameter wrong")
		}else if len(args[6]) <= 0{
			return  nil,errors.New("the address Parameter wrong")
		}
		//update argument
		fmt.Println("=== start init customer ===")
		res.CardId = cardid
		res.Name = name
		res.TelNo = telno
		res.Age = age
		res.Birthday = birthday
		res.Occupation = occupation
		res.Address = address
		str, err := json.Marshal(res)
		err = stub.PutState(cardid,str)
		if err != nil {
			return nil,errors.New("can't put into block ")
		}

		fmt.Println("- end update customer complete")
		return  nil,nil
	} else {
		return  nil, errors.New("Name not found")
	}

}
// ===================================================================================================================
	// update broker
// ===================================================================================================================
func  (t *SimpleChaincode) updateBroker(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	//define argument
	brkNo := args[0]
	brkName := args[1]
	if len(args[0]) <= 0 {
		return  nil,errors.New("Need  Broker Number paramenter ")
	}else if len(args[1]) <= 0{
		return  nil, errors.New("Need  Broker Name paramenter")
	}
	AsByteBrk,err := stub.GetState(brkNo)
	if err != nil {
		return nil, errors.New("Can't Get State of broker number from blockchain   ")
	}
	res := Broker{}
	json.Unmarshal(AsByteBrk,&res)
	// check and update broker
	if res.BrokerNo == brkNo {
		res.BrokerNo = brkNo
		res.BrokerName = brkName
		strJson,err  :=  json.Marshal(res)
		if err != nil{
			return  nil, errors.New("Can't make string to JSON")
		}
		err = stub.PutState(brkNo,strJson)
		if err != nil {
			return  nil, errors.New("Can't Put State to Blockchain ")
		}
		return AsByteBrk,nil
	}else {
		return  nil, errors.New("Broker not found")
	}
}
// ============================================================================================================================
// Init Customer - create a new customer, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) new_customer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error


	//   0       1       2     3
	// "asdf", "blue", "35", "bob"
	if len(args) != 7 {
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
	if len(args[2]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return nil, errors.New("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0{
		return nil , errors.New("5th argument must be a non-empty string")
	}
	if len(args[5]) <= 0{
		return nil, errors.New("6th argument must be a non-empty string")
	}
	if len(args[6]) <= 0{
		return nil, errors.New("7th argument must be a non-empty string")
	}
	cardid := args[0]
	name := args[1]
	telno := strings.ToLower(args[2])
	size, err := strconv.Atoi(args[3])
	if err != nil {
		return nil, errors.New("3rd argument must be a numeric string")
	}
	birthday := args[4]
	occupation := strings.ToLower(args[5])
	address := args[6]

	//check if customer already exists
	customerAsBytes, err := stub.GetState(cardid)
	if err != nil {
		return nil, errors.New("Failed to get customer name")
	}
	//jsonMap := make(map[string]interface{})
	res := Customer{}
	json.Unmarshal([]byte(customerAsBytes), &res)
	if res.CardId == cardid {
		fmt.Println("This customer arleady exists: " + cardid)
		fmt.Println(res)
		return nil, errors.New("This customer arleady exists") //all stop a customer by this name exists
	}
	res.CardId = cardid
	res.Name = name
	res.TelNo = telno
	res.Age = size
	res.Birthday = birthday
	res.Occupation = occupation
	res.Address = address
	//build the customer json string manually
	//str := `{"name": "` + name + `", "telno": "` + telno + `", "size": ` + strconv.Itoa(size) + `, "user": "` + user + `"}`
	//err = stub.PutState(name, []byte(str)) //store customer with id as key
	str, err := json.Marshal(res)
	err = stub.PutState(cardid, str)
	if err != nil {
		return nil, err
	}

	//get the customer index
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
//=====================================================================================================================
	//Add Nwe Broker
//====================================================================================================================
func (t *SimpleChaincode) newBroker(stub shim.ChaincodeStubInterface, args []string)  ([]byte, error){
	var err error
	//check parameter not null
	if len(args[0]) <= 0 {
		return nil,errors.New("Broker Number wrong Parameter")
	}else if len(args[0]) <= 0{
		return  nil,errors.New("Broker Name wrong Parameter")
	}
	brkno := args[0]
	brkname := args[1]
	res  := Broker{}
	AsByteBrk,err := stub.GetState(brkno)
	if err != nil {
		return nil, errors.New("Can't Get State from blockchain ")
	}
	json.Unmarshal(AsByteBrk,&res)

	if res.BrokerNo == brkno{
		return nil, errors.New("Broker Duplicate")
	}else {
		//add new Broker
		res.BrokerNo = brkno
		res.BrokerName = brkname
		strJson, err := json.Marshal(res)
		if err != nil {
			return nil, errors.New("Can't make bye to Json ")
		}
		err = stub.PutState(brkno,strJson)
		if err != nil {
			return nil, err
		}
		return  nil,nil
	}

}
//==============================================================================================================
	//update guarantee ID
//==============================================================================================================
func (t *SimpleChaincode) updateGuaranteeID(stub shim.ChaincodeStubInterface, args []string)	([]byte, error){
	var err error
	if len(args) < 5 {
		return nil,errors.New("No parameter require")
	}

	cardid := args[1]
	name := args[2]
	AsbyteGusrantee,err := stub.GetState(cardid)
	if err != nil{
		return  nil,errors.New("Can't Get State data")
	}
	res := guaranteeID{}
	json.Unmarshal(AsbyteGusrantee,&res)
	if res.CardID == cardid {
		res.Name = name
		fmt.Println("update success")
		strJson,err := json.Marshal(cardid)
		if err != nil {
			return nil,errors.New("Can't make to json ")
		}
		err  = stub.PutState(cardid,strJson)
		if err != nil{
			return  nil,errors.New("Can't put state to block chain ")
		}
		return []byte(res.Name),nil
	}else {
		return  nil , errors.New("Card id not found")
	}

}
//===========================================================================================================
	//update allow broker
//==========================================================================================================
func (t *SimpleChaincode) updateAllowBroker(stub shim.ChaincodeStubInterface, args []string)	([]byte, error){
	var err error
	//Assume consensus node allowed all broker
	if len(args) < 2{
		return nil, errors.New("Require Paremeter wrong")
	}
	cardid := args[0]
	guarantee_id := args[1]
	allowBrk := args[2]
	res_Cust_Brk := Customer{}
	AsBytecustomer,err :=  stub.GetState(cardid)
	if err != nil {
		return errors.New("")
	}else if len(cardid) == 0 {
		return nil,errors.New("Argument is NULL")
	}else if len(guarantee_id) == 0{
		return nil , errors.New("Argument is NULL")
	}else if len(allowBrk) == 0 {
		return  nil,errors.New("Argument is NULL")
	}else {
		json.Unmarshal(AsBytecustomer,&res_Cust_Brk)
		if res_Cust_Brk.CardId == cardid && res_Cust_Brk.GauranteeID == guarantee_id {
			res_Cust_Brk.AllowBroke = allowBrk
			strJson,err := json.Marshal(res_Cust_Brk)
			if err != nil{
				return nil ,errors.New("Can't make to JSON")
			}
			err = stub.PutState(cardid,strJson)
			if err != nil{
				return  nil,errors.New("Can't put state to block chain")
			}
		}else {
			return  nil ,errors.New("Card ID not found")
		}

		return  res_Cust_Brk.AllowBroke,nil
	}



}
// ============================================================================================================================
// Set User Permission on Customer
// ============================================================================================================================
func (t *SimpleChaincode) set_user(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//   0       1
	// "name", "bob"
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	fmt.Println("- start set user")
	fmt.Println(args[0] + " - " + args[1])
	customerAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	res := Customer{}
	json.Unmarshal(customerAsBytes, &res) //un stringify it aka JSON.parse()
	//res.User = args[1]                  //change the user

	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes) //rewrite the customer with id as key
	if err != nil {
		return nil, err
	}

	fmt.Println("- end set user")
	return nil, nil
}
