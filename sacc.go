/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"strconv"
	"encoding/json"
	// timestamp "github.com/golang/protobuf/ptypes/timestamp"
	// ptypes "github.com/golang/protobuf/ptypes"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/ledger/queryresult"
)

// SimpleAsset implements a simple chaincode to manage an asset
type SimpleAsset struct {
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Get the args from the transaction proposal
	args := stub.GetStringArgs()
	// if len(args) != 2 {
	// 	return shim.Error("Incorrect arguments. Expecting a key and a value")
	// }

	/// for Oracle blockchain: [0]= "init" at default, [1]=key, [2]=value
	if len(args) != 3 {
		return shim.Error("Incorrect arguments. Expecting a key and a value")
	}

	// Set up any variables or assets here by calling stub.PutState()

	// We store the key and the value on the ledger
	//err := stub.PutState(args[0], []byte(args[1]))

	// for Oracle blockchain
	err := stub.PutState(args[1], []byte(args[2]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
	}
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error
	if fn == "set" {
		result, err = set(stub, args)
	} else if fn == "get" {
		result, err = get(stub, args)
	} else if fn == "getHistoryOfState" {
		result, err = getHistoryOfState(stub, args)
	} else {
		return shim.Error("Invalid chaincode function name")
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success([]byte(result))
}

func getHistoryOfState(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("Must provide a key")
	}

	historyQueryIterator, err := stub.GetHistoryForKey(args[0])
	if err != nil {
		return "", fmt.Errorf("Error in fetching history: " + err.Error())
	}

	var resultModification *queryresult.KeyModification
	//counter := 0
	//resultJSON := "["

	type history struct {
		TxID		string		`json:"txID"`
		Value		int		`json:"value"`
		//Timestamp	string	`json:"timestamp"`
		IsDelete	bool		`json:"isDelete"`

	}

	type returnValues struct {
		NumberOfTransactions	int		`json:"numberOfTransactions"`
		AssetName				string	`json:"assetName"`
		History					[]history	`json:"history"`
	}

	response := returnValues {
		NumberOfTransactions: 0,
		AssetName: args[0],
	}

	for historyQueryIterator.HasNext() {
		resultModification, err = historyQueryIterator.Next()
		if err != nil {
			return "", fmt.Errorf("Error in reading history record" + err.Error())
		}

		val, _ := strconv.Atoi(string(resultModification.GetValue()))
		
		data := history {
			TxID: resultModification.GetTxId(),
			Value: val,
			IsDelete: resultModification.GetIsDelete(),
		}

		response.NumberOfTransactions ++
		response.History = append(response.History, data)

	}

	historyQueryIterator.Close()

	responseJSON, _ := json.Marshal(response)

	return string(responseJSON), nil
	
}


// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
func set(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	//return args[1], nil
	return "The state of the asset " + args[0] + " has been setup successfully with the value " + args[1], nil
}

// Get returns the value of the specified asset key
func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}

	//return string(value), nil

	type returnValues struct {
		Key 	string 	`json:"key"`
		Value 	string	`json:"value"`
	}

	response := returnValues{
		Key: args[0],
		Value : string(value),
	}

	responseJSON, _ := json.Marshal(response)

	return string(responseJSON), nil

}


// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
