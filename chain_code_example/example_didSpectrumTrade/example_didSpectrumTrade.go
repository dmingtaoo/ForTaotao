package main

import (
	cc "dpchain/core/contract/chainCodeSupport"
	"flag"
	"fmt"
)

var (
	ERROR_FUNCTION_ARGS = fmt.Errorf("unmatched arguments")
)

var (
	CONTRACT_NAME = "DID::SPECTRUM::TRADE"
)

var dataList [][]byte

func SetDataPipe(args [][]byte, ds cc.DperServicePipe) ([][]byte, error) {
	if len(args) != 2 {
		return nil, ERROR_FUNCTION_ARGS
	}
	str := string(args[1])
	if len(str) >= 4 && str[:4] == "DID:" {
		ds.UpdateStatus(args[0], args[1])
		result := [][]byte{[]byte("setDID succeed")}
		dataList = append(dataList, args[0],args[1])
		return result, nil
	} else {
		return nil, ERROR_FUNCTION_ARGS
	}
}

func GetDataPipe(args [][]byte, ds cc.DperServicePipe) ([][]byte, error) {
	if len(args) != 1 {
		return nil, ERROR_FUNCTION_ARGS
	}
	value, err := ds.GetStatus(args[0])
	if err != nil {
		return nil, err
	}
	result := [][]byte{value}
	return result, nil
}

func GetDataList(args [][]byte, ds cc.DperServicePipe) ([][]byte, error) {
	data := dataList
	return data, nil
}

func main() {
	local_pipe := flag.String("local_pipe", "", "")
	Dper_pipe := flag.String("dper_pipe", "", "")
	flag.Parse()
	funcMap := map[string]cc.ContractFuncPipe{
		"SetDataPipe": SetDataPipe,
		"GetDataPipe": GetDataPipe,
		"GetDataList": GetDataList,
	}
	err := cc.InstallContractPipe(CONTRACT_NAME, funcMap, *Dper_pipe)
	if err != nil {
		fmt.Print(err)
	} else {
		fmt.Print("install success")
	}
	cc.ContractExecutePipe(*Dper_pipe, *local_pipe, funcMap)
}

