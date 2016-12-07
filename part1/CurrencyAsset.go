package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var logger = shim.NewLogger("CLDChaincode")

//==============================================================================================================================
//	 Participant types - Each participant type is mapped to an integer which we use to compare to the value stored in a
//						 user's eCert
//==============================================================================================================================
//CURRENT WORKAROUND USES ROLES CHANGE WHEN OWN USERS CAN BE CREATED SO THAT IT READ 1, 2, 3, 4, 5
const AUTHORITY = "regulator"
const SUBSCRIBER = "subscriber"
const PRIVATE_ENTITY = "private"
const USER = "user"

//==============================================================================================================================
//	 Structure Definitions
//==============================================================================================================================
//	Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//				and other HyperLedger functions)
//==============================================================================================================================
type SimpleChaincode struct {
}

//==============================================================================================================================
//	CurrencyAsset - define the asset
//==============================================================================================================================

type CurrencyAsset struct {
	CurrId   string             `json:"currid"`
	Name     string             `json:"name"`
	CType    int                `json:"ctype"`
	Owner    string             `json:"owner"`
	Balances map[string]float64 `json:"balances"`
}

//==============================================================================================================================
//	wallet - define the asset
//==============================================================================================================================

//==============================================================================================================================
//	CurrencyAsset Holder - hold the currency
//==============================================================================================================================

type CurrencyHolder struct {
	IDs []string `json:"ids"`
}

//==============================================================================================================================
//	User_and_eCert - Struct for storing the JSON of a user and their ecert
//==============================================================================================================================

type User_and_eCert struct {
	Identity string `json:"identity"`
	eCert    string `json:"ecert"`
}

type CurrentBalance struct {
	Identity string `json:"identity"`
	Balance  float64 `json:"balance"`
}

//==============================================================================================================================
//	Init Function - Called when the user deploys the chaincode
//==============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	var currHolder CurrencyHolder

	bytes, err := json.Marshal(currHolder)

	if err != nil {
		return nil, errors.New("Error creating currHolder record")
	}

	err = stub.PutState("currHolder", bytes)

	for i := 0; i < len(args); i = i + 2 {
		t.addecert(stub, args[i], args[i+1])
	}

	return nil, nil
}

//==============================================================================================================================
//	 General Functions
//==============================================================================================================================
//	 getecert - Takes the name passed and calls out to the REST API for HyperLedger to retrieve the ecert
//				 for that user. Returns the ecert as retrived including html encoding.
//==============================================================================================================================
func (t *SimpleChaincode) getecert(stub shim.ChaincodeStubInterface, name string) ([]byte, error) {

	ecert, err := stub.GetState(name)

	if err != nil {
		return nil, errors.New("Couldn't retrieve ecert for user " + name)
	}

	return ecert, nil
}

//==============================================================================================================================
//	 addecert - Adds a new ecert and user pair to the table of ecerts
//==============================================================================================================================

func (t *SimpleChaincode) addecert(stub shim.ChaincodeStubInterface, name string, ecert string) ([]byte, error) {

	err := stub.PutState(name, []byte(ecert))

	if err == nil {
		return nil, errors.New("Error storing eCert for user " + name + " identity: " + ecert)
	}

	return nil, nil

}

//==============================================================================================================================
//	 get_caller - Retrieves the username of the user who invoked the chaincode.
//				  Returns the username as a string.
//==============================================================================================================================

func (t *SimpleChaincode) getusername(stub shim.ChaincodeStubInterface) (string, error) {

	username, err := stub.ReadCertAttribute("username")
	if err != nil {
		return "", errors.New("Couldn't get attribute 'username'. Error: " + err.Error())
	}
	return string(username), nil
}

//==============================================================================================================================
//	 check_affiliation - Takes an ecert as a string, decodes it to remove html encoding then parses it and checks the
// 				  		certificates common name. The affiliation is stored as part of the common name.
//==============================================================================================================================

func (t *SimpleChaincode) checkaffiliation(stub shim.ChaincodeStubInterface) (string, error) {
	affiliation, err := stub.ReadCertAttribute("role")
	if err != nil {
		return "", errors.New("Couldn't get attribute 'role'. Error: " + err.Error())
	}
	return string(affiliation), nil

}

//==============================================================================================================================
//	 get_caller_data - Calls the getecert and check_role functions and returns the ecert and role for the
//					 name passed.
//==============================================================================================================================

func (t *SimpleChaincode) getcallerdata(stub shim.ChaincodeStubInterface) (string, string, error) {

	user, err := t.getusername(stub)

	// if err != nil { return "", "", err }

	// ecert, err := t.getecert(stub, user);

	// if err != nil { return "", "", err }

	affiliation, err := t.checkaffiliation(stub)

	if err != nil {
		return "", "", err
	}

	return user, affiliation, nil
}

//==============================================================================================================================
//	 retrievecurrencyassetc - Gets the state of the data at currId in the ledger then converts it from the stored
//					JSON into the CurrencyAsset struct for use in the contract. Returns the CurrencyAsset struct.
//					Returns empty v if it errors.
//==============================================================================================================================
func (t *SimpleChaincode) retrievecurrencyassetinfo(stub shim.ChaincodeStubInterface, currId string) (CurrencyAsset, error) {

	var currencyAsset CurrencyAsset

	bytes, err := stub.GetState(currId)

	if err != nil {
		fmt.Printf("RETRIEVE_CURR_ASSET: Failed to invoke currId: %s", err)
		return currencyAsset, errors.New("RETRIEVE_CURR_ASSET: Error retrieving CurrencyAsset with currId = " + currId)
	}

	err = json.Unmarshal(bytes, &currencyAsset)

	if err != nil {
		fmt.Printf("RETRIEVE_CURR_ASSET: Corrupt CurrencyAsset with record "+string(bytes)+": %s", err)
		return currencyAsset, errors.New("RETRIEVE_CURR_ASSET: Corrupt CurrencyAsset record" + string(bytes))
	}

	return currencyAsset, nil
}

func (t *SimpleChaincode) retrievebalance (stub shim.ChaincodeStubInterface, currencyAsset CurrencyAsset, userid string) ([]byte, error) {

       // get the currencyAssetInfo

      var balance = currencyAsset.Balances[userid] 


      var cb CurrentBalance 
      cb.Identity = userid
      cb.Balance = balance

      bytes,err := json.Marshal(cb)

      return bytes,err

}






//==============================================================================================================================
// savechanges - Writes to the ledger the CurrencyAsset struct passed in a JSON format. Uses the shim file's
//				  method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) savechanges(stub shim.ChaincodeStubInterface, currAss CurrencyAsset) (bool, error) {

	bytes, err := json.Marshal(currAss)

	if err != nil {
		fmt.Printf("SAVE_CHANGES: Error converting CurrencyAsset record: %s", err)
		return false, errors.New("Error converting CurrencyAsset record")
	}

	err = stub.PutState(currAss.CurrId, bytes)

	if err != nil {
		fmt.Printf("SAVE_CHANGES: Error storing CurrencyAsset record: %s", err)
		return false, errors.New("Error storing CurrencyAsset record")
	}

	return true, nil
}


//==============================================================================================================================
//	 Router Functions
//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		  initial arguments passed to other things for use in the called function e.g. name -> ecert
//==============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	caller, caller_affiliation, err := t.getcallerdata(stub)

	if err != nil {
		return nil, errors.New("Error retrieving caller information")
	}

	if function == "createCurrencyAsset" {
		return t.createcurrencyAsset(stub, caller, caller_affiliation, args[0], args[1], args[2],args[3])
	} else if function == "ping" {
		return t.ping(stub)
	} else { // If the function is not a create then there must be a car so we need to retrieve the car.
		argPos := 1

		currAsset, err := t.retrievecurrencyassetinfo(stub, args[argPos])

		if err != nil {
			fmt.Printf("INVOKE: Error retrieving currencyAsset: %s", err)
			return nil, errors.New("Error retrieving currencyAsset")
		}
		return json.Marshal (currAsset)

	}
	return nil, errors.New("Error function not implemented");
}

//=================================================================================================================================
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.
//=================================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	caller, caller_affiliation, err := t.getcallerdata(stub)
	if err != nil {
		fmt.Printf("QUERY: Error retrieving caller details", err)
		return nil, errors.New("QUERY: Error retrieving caller details: " + err.Error())
	}

	logger.Debug("function: ", function)
	logger.Debug("caller: ", caller)
	logger.Debug("affiliation: ", caller_affiliation)

	if function == "getCurrencyAssetInfo" {
		if len(args) != 1 {
			fmt.Printf("Incorrect number of arguments passed")
			return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}
		currAsset, err := t.retrievecurrencyassetinfo(stub, args[0])
		if err != nil {
			fmt.Printf("QUERY: Error retrieving currencyAsset: %s", err)
			return nil, errors.New("QUERY: Error retrieving currencyAsset " + err.Error())
		}
		return json.Marshal (currAsset)
	} else if function == "getecert" {
		return t.getecert(stub, args[0])
	} else if function == "ping" {
		return t.ping(stub)
	} else if function == "getBalance" {

		// get the currencyAsset first
		 currencyAsset, err := t.retrievecurrencyassetinfo(stub, args[0])
		if  err != nil {
		     fmt.Printf("QUERY: Error retrieving currencyAsset: %s", err)
		     return nil, errors.New("QUERY: Error retrieving currencyAsset " + err.Error())
		}

		v, err := t.retrievebalance(stub, currencyAsset,args[1])
		if  err != nil {
                     fmt.Printf("QUERY: Error retrieving currencyAsset: %s", err)
                     return nil, errors.New("QUERY: Error retrieving currencyAsset " + err.Error())
                }

		return  v, nil

	}
	return nil, errors.New("Received unknown function invocation " + function)

}

//=================================================================================================================================
//	 Ping Function
//=================================================================================================================================
//	 Pings the peer to keep the connection alive
//=================================================================================================================================
func (t *SimpleChaincode) ping(stub shim.ChaincodeStubInterface) ([]byte, error) {
	return []byte("Hello, world!"), nil
}

//=================================================================================================================================
//	 Create Function
//=================================================================================================================================
//	 Create CurrencyAsset - Creates the initial JSON for the CurrencyAsset and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) createcurrencyAsset(stub shim.ChaincodeStubInterface, caller string, caller_affiliation string, pcurrId string, pname string, ptype string, powner string) ([]byte, error) {
	var currencyAsset CurrencyAsset

	// prepare the json string fields
    CurrId := "\"currid\":\"" + pcurrId + "\", " // Variables to define the JSON
	Name := "\"name\":\""+ pname + "\", "
	Owner := "\"owner\":\"" + powner + "\", "
	Ctype := "\"ctype\":" + ptype  

	// prepare the json record
	currencyAsset_json := "{" + CurrId+Name+Owner+Ctype + "}" // Concatenates the variables to create the total JSON object


	err := json.Unmarshal([]byte(currencyAsset_json), &currencyAsset) // Convert the JSON defined above into a vehicle object for go


	if err != nil {
		return nil, errors.New("Invalid JSON object")
	}

	//If not an error then a record exists so cant create
	// new  with this currId as it must be unique

	record, err := stub.GetState(currencyAsset.CurrId)
	if record != nil {
		return nil, errors.New("CurrencyAsset already exists")
	}

	if caller_affiliation != AUTHORITY { // Only the regulator can create a new

		return nil, errors.New(fmt.Sprintf("Permission Denied. create_CurrencyAsset. %v === %v", caller_affiliation, AUTHORITY))

	}

	_, err = t.savechanges(stub, currencyAsset)

	if err != nil {
		fmt.Printf("CREATE_: Error saving changes: %s", err)
		return nil, errors.New("Error saving changes")
	}

	bytes, err := stub.GetState("currHolder")

	if err != nil {
		return nil, errors.New("Unable to get currHolder")
	}

	var currHolder CurrencyHolder

	err = json.Unmarshal(bytes, &currHolder)

	if err != nil {
		return nil, errors.New("Corrupt CurrencyHolder record")
	}

	currHolder.IDs = append(currHolder.IDs, pcurrId)

	bytes, err = json.Marshal(currHolder)

	if err != nil {
		fmt.Print("Error creating CurrencyHolder record")
	}

	err = stub.PutState("currHolder", bytes)

	if err != nil {
		return nil, errors.New("Unable to put the state")
	}

	return nil, nil

}

//=================================================================================================================================
//	 Main - main - Starts up the chaincode
//=================================================================================================================================
func main() {

	err := shim.Start(new(SimpleChaincode))

	if err != nil {
		fmt.Printf("Error starting Chaincode: %s", err)
	}
}



