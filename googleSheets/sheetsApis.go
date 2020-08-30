package googleSheets

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	config "github.com/bhambri94/ig-reports/configs"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

var srv *sheets.Service
var spreadsheetId string

// Retrieve a token, saves the token, then returns the generated client.
func getClient() *sheets.Service {
	b, err := ioutil.ReadFile("googleSheets/secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "googleSheets/token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}

	srv, err := sheets.New(config.Client(context.Background(), tok))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	return srv
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func Read(readRange string) {
	if srv == nil {
		srv = getClient()
	}
	spreadsheetId = config.Configurations.SpreadsheetID
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		for _, row := range resp.Values {
			fmt.Printf("%s, %s\n", row[0], row[0])
		}
	}
}

func BatchWrite(SheetName string, value [][]interface{}) {
	if srv == nil {
		srv = getClient()
	}
	spreadsheetId = config.Configurations.SpreadsheetID
	rb := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED",
	}
	rb.Data = append(rb.Data, &sheets.ValueRange{
		Range:  SheetName,
		Values: value,
	})
	fmt.Println("Writing data to Google Sheets with data")
	_, err := srv.Spreadsheets.Values.BatchUpdate(spreadsheetId, rb).Context(context.Background()).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
	} else {
		fmt.Println("Data has been successfully pushed to Google Sheet")
	}
}

type BatchGetResponse struct {
	SpreadsheetID string `json:"spreadsheetId"`
	ValueRanges   []struct {
		MajorDimension string     `json:"majorDimension"`
		Range          string     `json:"range"`
		Values         [][]string `json:"values"`
	} `json:"valueRanges"`
}

func BatchGet(SheetRange string) [][]string {
	ranges := []string{SheetRange}
	if srv == nil {
		srv = getClient()
	}
	spreadsheetId = config.Configurations.SpreadsheetID
	resp, err := srv.Spreadsheets.Values.BatchGet(spreadsheetId).Ranges(ranges...).Context(context.Background()).Do()
	if err != nil {
		log.Fatal(err)
	}

	a, _ := resp.MarshalJSON()
	fmt.Printf(string(a))

	var batchGetResponse BatchGetResponse
	error := json.Unmarshal(a, &batchGetResponse)
	if error != nil {
		panic(error)
	}
	return batchGetResponse.ValueRanges[0].Values
}

func BatchAppend(SheetName string, value [][]interface{}) {
	ctx := context.Background()
	if srv == nil {
		srv = getClient()
	}

	spreadsheetId := config.Configurations.SpreadsheetID
	valueInputOption := "RAW"
	insertDataOption := "INSERT_ROWS"
	rb := &sheets.ValueRange{
		Range:  SheetName,
		Values: value,
	}

	resp, err := srv.Spreadsheets.Values.Append(spreadsheetId, SheetName, rb).ValueInputOption(valueInputOption).InsertDataOption(insertDataOption).Context(ctx).Do()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#v\n", resp)
}

func ClearSheet(SheetName string) {
	readRange := SheetName
	var itt bool
	if srv == nil {
		srv = getClient()
	}
	spreadsheetId = config.Configurations.SpreadsheetID
	_, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		fmt.Printf("Unable to retrieve data from sheet name: %v", err)
		fmt.Println()
		itt = true
	}

	if itt {
		fmt.Println("Creating new sheet with sheetname: " + SheetName)
		req := sheets.Request{
			AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{
					Title: SheetName,
				},
			},
		}
		rbb := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{&req},
		}
		_, err := srv.Spreadsheets.BatchUpdate(spreadsheetId, rbb).Context(context.Background()).Do()
		if err != nil {
			log.Fatal(err)
		}

	} else {
		fmt.Println("Clearing the sheet's old data and pulling data for new Day")
		ranges := SheetName
		rb := &sheets.ClearValuesRequest{}
		_, err := srv.Spreadsheets.Values.Clear(spreadsheetId, ranges, rb).Context(context.Background()).Do()
		if err != nil {
			log.Fatal(err)
		}
	}
}
