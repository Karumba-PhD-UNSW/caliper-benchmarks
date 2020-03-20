package main

import (
	"fmt"
	"log"
	"strconv"
	"encoding/json"
	"encoding/base64"
	"crypto/sha256"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type SmartContract struct{
}

type Asset struct{
	Id string
	Type string
	Owner string
	Value string
	Nspace string
}

type Trade struct{
	Id string
	Input Asset
	Output Asset
}

const ERROR_SYSTEM = "{\"code\":300, \"reason\": \"system error: %s\"}"
const ERROR_WRONG_FORMAT = "{\"code\":301, \"reason\": \"command format is wrong, "
const ERROR_ASSET_EXISTING = "{\"code\":302, \"reason\": \"asset already exists\"}"
const ERROR_ASSET_ABNORMAL = "{\"code\":303, \"reason\": \"abnormal asset\"}"
const ERROR_VALUE_NOT_ENOUGH = "{\"code\":304, \"reason\": \"asset's value is not enough\"}"


// Init initializes chaincode
// ===========================
func (t *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	// nothing to do
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	myarg,_ := json.Marshal(args)
	log.Println("Invoke: Function: %s args: %s\n", function, string(myarg))

	if function == "createAsset" {
		return t.createAsset(stub, args)
	}
	if function == "queryAsset" {
		return t.queryAsset(stub, args)
	}
	if function == "deleteAsset" {
		return t.deleteAsset(stub, args)
	}
	if function == "p2pTrade" {
		return t.p2pTrade(stub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (t *SmartContract) createAsset(stub shim.ChaincodeStubInterface, args []string)  pb.Response {
	if len(args) != 1 {
		return shim.Error(ERROR_WRONG_FORMAT)
	}

	var req Asset
	err := json.Unmarshal([]byte(args[0]), &req)
	if err != nil {
		return shim.Error(ERROR_WRONG_FORMAT+"should be 'asset {id:string, type:string, owner:string, value:string, nspace:string '\"}")
	}

	// calculate the hash value as identity
	hash := sha256.Sum256([]byte(req.Id))
	id   := base64.URLEncoding.EncodeToString(hash[:])
	asset := req
	i,_  := json.Marshal(asset)

	log.Println("createAsset: saving: ", string(i))
	
	// check if asset already exists
	state, err := stub.GetState(id)
	if state != nil && err == nil {
		return shim.Error(ERROR_ASSET_EXISTING);
	}


	err = stub.PutState(id, []byte(i))
	if err != nil {
		fmt.Printf(ERROR_SYSTEM, err.Error())
		return shim.Error(err.Error())
	}

	indexName := "type~owner"
	typeOwnerIndexKey, err := stub.CreateCompositeKey(indexName, []string{req.Type, req.Owner})
	if err != nil {
		return shim.Error(err.Error())
	}

	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the marble.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(typeOwnerIndexKey, value)

	return shim.Success([]byte(id));

}

func (t *SmartContract) queryAsset(stub shim.ChaincodeStubInterface, args []string)  pb.Response {
	if len(args) != 1 {
		return shim.Error("Wrong format, should be 'query id'")
	}

	hash  := sha256.Sum256([]byte(args[0]))
	id    := base64.URLEncoding.EncodeToString(hash[:])
	stat,err := stub.GetState(id)
	if err != nil {

		jsonResp := "{\"Error\":\"Failed to get state for " +id + "\"}"
		return shim.Error(jsonResp)
	} 
	log.Println("queryAsset: state: ", string(stat))

	var item Asset
	err = json.Unmarshal(stat, &item)
	if err != nil {
		fmt.Printf("unknown state value %v \n", string(stat))
		return shim.Error(err.Error())
	}

	r,_  := json.Marshal(item)
	return shim.Success(r)
}

func (t *SmartContract) deleteAsset(stub shim.ChaincodeStubInterface, args []string)  pb.Response {
	var jsonResp string
	var assetJSON Asset
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	
	hash  := sha256.Sum256([]byte(args[0]))
	assetId    := base64.URLEncoding.EncodeToString(hash[:])
	// to maintain the type~value index, we need to read the asset first and get its type
	valAsbytes, err := stub.GetState(assetId) //get the asset from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + assetId + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Asset does not exist: " + assetId + "\"}"
		return shim.Error(jsonResp)
	}

	err = json.Unmarshal([]byte(valAsbytes), &assetJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + assetId + "\"}"
		return shim.Error(jsonResp)
	}

	err = stub.DelState(assetId) //remove the asset from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// maintain the index
	indexName := "type~owner"
	typeOwnerIndexKey, err := stub.CreateCompositeKey(indexName, []string{assetJSON.Type, assetJSON.Owner})
	if err != nil {
		return shim.Error(err.Error())
	}

	//  Delete index entry to state.
	err = stub.DelState(typeOwnerIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}
	return shim.Success(nil)
}

func (t *SmartContract) p2pTrade(stub shim.ChaincodeStubInterface, args []string)  pb.Response {
	//check args formart
	if len(args) != 1 {
		return shim.Error(ERROR_WRONG_FORMAT)
	}

	//unmarshal trade
	var req Trade
	err := json.Unmarshal([]byte(args[0]), &req)
	if err != nil {
		log.Println("p2pTrade: Unmarshal Error: ", err.Error())
		return shim.Error(ERROR_WRONG_FORMAT+"should be 'asset {id:string, type:string, owner:string, value:string, nspace:string '\"}")
	}

	log.Println("p2pTrade: request: ", string(req.Input.Id))
	var input, output Asset
	input = req.Input
	output = req.Output

	//prepare input key
	hash := sha256.Sum256([]byte(input.Id))
	inputId   := base64.URLEncoding.EncodeToString(hash[:])
	//prepare ouput key
	hash = sha256.Sum256([]byte(output.Id))
	outputId   := base64.URLEncoding.EncodeToString(hash[:])

	//get input asset state
	stat,err := stub.GetState(inputId)
	if err != nil {
		log.Println("p2pTrade: GetState Error: ", err.Error())
		jsonResp := "{\"Error\":\"Failed to get state for " +inputId + "\"}"
		return shim.Error(jsonResp)
	} 

	log.Println("p2pTrade: state: ", string(stat))

	//unmarshal state to asset 
	var item Asset
	err = json.Unmarshal(stat, &item)
	if err != nil {
		log.Println("unknown state value %v \n", string(stat))

		return shim.Error(err.Error())
	}

	//compute trade
	var old, new, val int
	old, err = strconv.Atoi(input.Value)
	if err != nil {
		log.Println("p2pTrade: Atoi Error: ", err.Error())
		return shim.Error(err.Error())
	}
	new, err = strconv.Atoi(output.Value)
	if err != nil {
		log.Println("p2pTrade: Atoi Error: ", err.Error())
		return shim.Error(err.Error())
	}

	//if not sufficient fund end trade
	if old < new {
		fmt.Println("- end Trade (failed-not enough value)")
		jsonResp := "{\"Error\":\" insufficient value for transfer " +item.Value + "\"}"
		return shim.Error(jsonResp)

	} 

	//prepare exchange
	val = old - new
	item.Value = strconv.Itoa(val)

	//marshal input and outpu
	i,_  := json.Marshal(item)
	j,_  := json.Marshal(output)

	//check if cross-channel trade
	if input.Nspace == output.Nspace {
		log.Println("Trade: P2P trade within same channel")

		// check if output asset already exists
		state, err := stub.GetState(outputId)
		if state != nil && err == nil {
			return shim.Error(ERROR_ASSET_EXISTING);
		}

		//create output as new asset
		err = stub.PutState(outputId, []byte(j))
		if err != nil {
			fmt.Printf(ERROR_SYSTEM, err.Error())
			return shim.Error(err.Error())
		}

		//update asset value
		err = stub.PutState(inputId, []byte(i))
		if err != nil {
			fmt.Printf(ERROR_SYSTEM, err.Error())
			return shim.Error(err.Error())
		}

	} else {
		log.Println("Trade: Relational trade with cross-channel exchanges")
		var channel string

		// invoke transfer in output channel
		chainCodeArgs := util.ToChaincodeArgs("createAsset",string(j))
		if input.Nspace == "myasset" {
			channel = "myChannel"

		} else {
			channel = "yourChannel"
		}
		response := stub.InvokeChaincode(output.Nspace, chainCodeArgs, channel)
		if response.Status != shim.OK {
           return shim.Error(response.Message)
        }


		//update asset value
		err = stub.PutState(inputId, []byte(i))
		if err != nil {
			fmt.Printf(ERROR_SYSTEM, err.Error())
			return shim.Error(err.Error())
		}

	}


	fmt.Println("- end Trade (success)")
	return shim.Success(nil)
}

func  main()  {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error starting chaincode: %v \n", err)
	}

}