package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/jessevdk/go-flags"
)

type RespVar struct {
	Name     string `json:"name"`
	JsonPath string `json:"jsonPath"`
	ApiType  string `json:"apiType"`
}

type WorkStep struct {
	ApiName                string                 `json:"apiName"`
	ApiEndpoint            string                 `json:"apiEndpoint"`
	ApiType                string                 `json:"apiType"`
	ExpectedResponseStatus string                 `json:"expectedResponseStatus"`
	ApiIndex               int                    `json:"apiIndex"`
	ApiPayload             map[string]interface{} `json:"apiPayload"`
	RespVars               []RespVar              `json:"respVar"`
}

type Results struct {
	VersaPostStagingTemplate []WorkStep
}

var envVar = make(map[string]string)

func initInputParams(envFileName string) {

	envFile, err := os.Open(envFileName)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), ":")
		if len(s) == 2 {
			envVar[strings.TrimSpace(s[0])] = strings.TrimSpace(s[1])
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	envFile.Close()
}

func printEnv() {
	for key, value := range envVar {
		fmt.Println(key,":",value)
	}
}

func replaceVar(inStr string) string {
	tmpStr := inStr
	for key, value := range envVar {
		old := "$*" + key + "$"
		tmpStr = strings.ReplaceAll(tmpStr, old, value)
	}
	return tmpStr
}

type Options struct {
	EnvFile string `short:"e" long:"env" required:"true" description:"environment file"`
	TemplateFile string `short:"t" long:"template" required:"true" description:"template file"`
	Server string `short:"s" long:"server" required:"true" description:"server name or IP address"`
	Port int `short:"p" long:"port" required:"true" description:"server port"`
}

var options Options

var parser = flags.NewParser(&options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}

	initInputParams(options.EnvFile)
	processTemplate(options.TemplateFile, options.Server, options.Port)
}


func processTemplate(tempFile string, server string, port int) {
		jsonFile, err := os.Open(tempFile)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("reading " + tempFile + " ...")
	
		byteValue, _ := ioutil.ReadAll(jsonFile)
		result := Results{}
		if err := json.Unmarshal(byteValue, &result); err != nil {
			panic(err)
		}
	
		jsonFile.Close()
	
		for i := 0; i < len(result.VersaPostStagingTemplate); i++ {
		// for i := 0; i < 4; i++ {
			fmt.Println("\nexecuting step", i+1, "of", len(result.VersaPostStagingTemplate))
			execWorkStep(result.VersaPostStagingTemplate[i], server, port)
		}
}

func execWorkStep(workStep WorkStep, server string, port int) {
	if workStep.ApiType == "POST" {
		execPostApi(workStep, server, port)
	} else if workStep.ApiType == "GET" {
		execGetApi(workStep, server, port)
	} else if workStep.ApiType == "PUT" {
		execPutApi(workStep, server, port)
	} else {
		log.Print("unsupported API type " + workStep.ApiType)
		return
	}
}

func execGetApi(workStep WorkStep, server string, port int) {
	protocol := "http"
	url := protocol + "://" + server + ":" + strconv.Itoa(port) + "/" + workStep.ApiEndpoint
	url = replaceVar(url)
	fmt.Println("[GET] " + url)

	res, err := http.Get(url)
	if err != nil {
		log.Print(err.Error())
		return
	}

	expRespCode, err := strconv.Atoi(workStep.ExpectedResponseStatus)
	if err != nil {
		log.Print("Error !!!" + err.Error())
		return
	}

	if res.StatusCode != expRespCode {
		log.Print("Error !!! response code is not " + workStep.ExpectedResponseStatus)
		return
	} else {
		log.Println("received success response", res.StatusCode, "from GET API")
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
	}
	fmt.Printf("[Response Body] %s", responseData)

	if workStep.RespVars != nil {
		for i := 0; i < len(workStep.RespVars); i++ {
			value := gjson.Get(string(responseData), workStep.RespVars[i].JsonPath)
			envVar[strings.TrimSpace(workStep.RespVars[i].Name)] = strings.TrimSpace(value.String())
		}
	}
}

func execPostApi(workStep WorkStep, server string, port int) {
	protocol := "http"
	url := protocol + "://" + server + ":" + strconv.Itoa(port) + "/" + workStep.ApiEndpoint
	url = replaceVar(url)
	fmt.Println("[POST] " + url)

	jsonStr, err := json.Marshal(workStep.ApiPayload)
	if err != nil {
		log.Print(err.Error())
		return
	}

	temp := replaceVar(string(jsonStr))
	fmt.Printf("[Request Body] %s\n", temp)
	jsonStr = []byte(temp)

	res, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Print(err.Error())
		return
	}

	expRespCode, err := strconv.Atoi(workStep.ExpectedResponseStatus)
	if err != nil {
		log.Print("Error !!!" + err.Error())
		return
	}

	if res.StatusCode != expRespCode {
		log.Print("Error !!! response code is not " + workStep.ExpectedResponseStatus)
		return
	} else {
		log.Println("received success response", res.StatusCode, "from POST API")
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Printf("[Response Body] %s\n", responseData)

}

func execPutApi(workStep WorkStep, server string, port int) {
	protocol := "http"
	url := protocol + "://" + server + ":" + strconv.Itoa(port) + "/" + workStep.ApiEndpoint
	url = replaceVar(url)
	fmt.Println("[PUT] " + url)

	jsonStr, err := json.Marshal(workStep.ApiPayload)
	if err != nil {
		log.Print(err.Error())
		return
	}

	temp := replaceVar(string(jsonStr))
	fmt.Printf("[Request Body] %s\n", temp)
	jsonStr = []byte(temp)

	// set the HTTP method, url, and request body
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}

	// set the request header Content-Type for json
	client := &http.Client{}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	expRespCode, err := strconv.Atoi(workStep.ExpectedResponseStatus)
	if err != nil {
		log.Print("Error !!!" + err.Error())
		return
	}

	if resp.StatusCode != expRespCode {
		log.Print("Error !!! response code is not " + workStep.ExpectedResponseStatus)
		return
	} else {
		log.Println("received success response", resp.StatusCode, "from PUT API")
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Printf("[Response Body] %s\n", responseData)
}