package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func getCustLANVrfDetails(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n[GET] %s", r.URL)
	b, _ := ioutil.ReadFile("interfaces.json")

	rawIn := json.RawMessage(string(b))
	var objmap map[string]*json.RawMessage
	err := json.Unmarshal(rawIn, &objmap)
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println(objmap)
	jsonString, _ := json.Marshal(objmap)
	fmt.Printf("\n[Response]%s\n", jsonString)

	json.NewEncoder(w).Encode(objmap)
}

func ConfigPostStageTemplate(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n[POST] %s", r.URL)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n[Body]%s\n", b)
	// json.NewEncoder(w).Encode(objmap)
	json.NewEncoder(w)
	// w.WriteHeader(http.StatusOK)
}

func CreateTviInterface(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n[POST] %s", r.URL)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n[Body]%s\n", b)
	// json.NewEncoder(w).Encode(objmap)
	json.NewEncoder(w)
	// w.WriteHeader(http.StatusOK)
}

func addTunnelInterface(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n[PUT] %s", r.URL)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n[Body]%s\n", b)
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/vnms/sdwan/workflow/templates/template", ConfigPostStageTemplate).Methods("POST")
	router.HandleFunc("/api/config/devices/template/Post-Staging-Template/config/interfaces", CreateTviInterface).Methods("POST")
	router.HandleFunc("/api/config/devices/template/PostStagingTemplate/config/routing-instances/routing-instance/boa_248230044-LAN-VR/interfaces", getCustLANVrfDetails).Methods("GET")
	router.HandleFunc("/api/config/devices/template/PostStagingTemplate/config/routing-instances/routing-instance/boa_248230044-LAN-VR/interfaces", addTunnelInterface).Methods("PUT")

	log.Fatal(http.ListenAndServe(":8000", router))
}
