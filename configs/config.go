package configs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

type Configs struct {
	SpreadsheetID       string `json:"SpreadsheetID"`
	SheetNameWithRange  string `json:"SheetNameWithRange"`
	ResearchJRSheetName string `json:"ResearchJRSheetName"`
	SessionId           string `json:"SessionId"`
}

var (
	Configurations = Configs{}
)

func SetConfig() {
	input, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	error := json.Unmarshal(input, &Configurations)
	if error != nil {
		fmt.Println("Config file is missing in root directory")
		panic(error)
	} else {
		fmt.Println("Follwing values has been picked from config values:")
		fmt.Println(Configurations)
	}
}
