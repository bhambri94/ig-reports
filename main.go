package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bhambri94/ig-reports/configs"
	"github.com/bhambri94/ig-reports/googleSheets"
	"github.com/bhambri94/ig-reports/ig"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

var (
	logger, _ = zap.NewProduction()
	sugar     = logger.Sugar()
)

func main() {
	configs.SetConfig()
	sugar.Infof("starting ig-reports app server...")
	defer logger.Sync()

	router := fasthttprouter.New()
	router.GET("/v1/get/ig/report/username=:USERNAME/SessionID=:SessionID", handleSaveIGReportToSheetsNew)
	router.GET("/v1/get/igr/database-backup", handleIGRDatabaseBackup)
	router.GET("/v1/session/check", handleSessionIDsChecker)
	router.GET("/v1/follower/count/SessionID=:SessionID", handleLatestFollowerCount)
	router.GET("/v1/get/nos/search/SessionID=:SessionID", handleNOSSearchSetup1)
	router.GET("/v2/get/nos/search/SessionID=:SessionID", handleNOSSearchSetupLatest)
	router.GET("/v3/get/nos/search/SessionID=:SessionID", handleNOSSearchSetup3)
	router.GET("/v4/get/nos/search/SessionID=:SessionID", handleNOSSearchSetup4)
	router.GET("/v5/get/nos/search/SessionID=:SessionID", handleNOSSearchSetup5)
	router.GET("/v6/get/nos/search/SessionID=:SessionID", handleNOSSearchSetup6)
	router.GET("/v7/get/nos/search/SessionID=:SessionID", handleNOSSearchSetup7)
	router.GET("/v8/get/nos/search/SessionID=:SessionID", handleNOSSearchSetup8)
	router.GET("/v9/get/nos/search/SessionID=:SessionID", handleNOSSearchSetup9)
	router.GET("/v10/get/nos/search/SessionID=:SessionID", handleNOSSearchSetup10)
	router.GET("/v1/get/account/database/username=:USERNAME/SessionID=:SessionID", handleIGRDatabaseBackup)
	router.GET("/v1/get/ig/research/username=:USERNAME/LatestFollowerCount=:LatestFollowerCount/MinFollower=:MinFollower/MaxFollower=:MaxFollower/MinN=:MinN/MinNStar=:MinNStar/NDelta=:NDelta/SessionID=:SessionID", handleSaveIGResearchToSheets)
	router.GET("/v1/get/ig/nos/username=:USERNAME/LatestFollowerCount=:LatestFollowerCount/MinFollower=:MinFollower/MaxFollower=:MaxFollower/MinN=:MinN/MinNStar=:MinNStar/NDelta=:NDelta/SessionID=:SessionID", handleSaveIGResearchToSheets)
	log.Fatal(fasthttp.ListenAndServe(":3003", router.Handler))
}

func handleLatestFollowerCount(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a latest follower count request from paperclip!")
	var finalValues [][]interface{}
	SessionID := ctx.UserValue("SessionID")
	if SessionID != nil {
		temp := SessionID.(string)
		temp = temp[1 : len(temp)-1]
		SessionID = temp
	}
	currentValueInSheets := googleSheets.BatchGet(configs.Configurations.FollowingCountSheetName)
	iter1 := 0
	for iter1 < len(currentValueInSheets) {
		var row []interface{}
		if len(currentValueInSheets[iter1]) > 0 {
			time.Sleep(2 * time.Second)
			UserID, _, _ := ig.GetUserIDAndFollower(currentValueInSheets[iter1][0], SessionID.(string))
			LatestFollowingCount := ig.GetLatestFollowingCount(UserID, SessionID.(string))
			row = append(row, currentValueInSheets[iter1][0], LatestFollowingCount)
			fmt.Println(row)
		}
		finalValues = append(finalValues, row)
		fmt.Println(finalValues)
		iter1++
	}
	if len(finalValues) > 0 {
		googleSheets.ClearSheet(configs.Configurations.FollowingCountSheetName)
		googleSheets.BatchWrite(configs.Configurations.FollowingCountSheetName, finalValues)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Google Sheet Updated"))
		sugar.Infof("calling session cookie checker success!")
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, not able to fetch data "))
		sugar.Infof("calling session cookie checker failure!")
	}
}

func handleSessionIDsChecker(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a session id checker request to Google Sheets!")
	var finalValues [][]interface{}
	loc, _ := time.LoadLocation("Europe/Rome")
	currentTime := time.Now().In(loc)
	Time := currentTime.Format("2006-01-02")
	currentValueInSheets := googleSheets.BatchGet(configs.Configurations.SessionIDSheetName + "!A3:D500")

	iter1 := 0
	for iter1 < len(currentValueInSheets) {
		var row []interface{}
		if len(currentValueInSheets[iter1]) > 0 {
			fmt.Println(currentValueInSheets[iter1][0])
			time.Sleep(2 * time.Second)
			status := ig.SessionIDChecker(currentValueInSheets[iter1][0])
			row = append(row, status, Time)
			fmt.Println(row)

		}
		finalValues = append(finalValues, row)
		fmt.Println(finalValues)
		iter1++
	}
	if len(finalValues) > 0 {
		googleSheets.BatchWrite(configs.Configurations.SessionIDSheetName+"!E3:F500", finalValues)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Google Sheet Updated"))
		sugar.Infof("calling session cookie checker success!")
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, not able to fetch data "))
		sugar.Infof("calling session cookie checker failure!")
	}
}

func handleNOSSearchSetup(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a NOS Search request to Google Sheets!")
	SearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearchSheetName + "!A2:I2")
	fmt.Println(SearchQueryFromNOS)
	var nosSearchFinalValues [][]interface{}
	var nosDashboardFinalValues [][]interface{}
	var nosLatestFollowerCountFinalValues [][]interface{}
	var MinFollower string
	var MaxFollower string
	var MinN string
	var MinNStar string
	var NDelta string

	if len(SearchQueryFromNOS) == 1 {
		if len(SearchQueryFromNOS[0]) > 4 {
			MinFollower = SearchQueryFromNOS[0][3]
			MaxFollower = SearchQueryFromNOS[0][5]
			MinFollower = strings.Replace(MinFollower, ",", "", -1)
			MaxFollower = strings.Replace(MaxFollower, ",", "", -1)
			if len(SearchQueryFromNOS[0]) > 6 {
				MinN = SearchQueryFromNOS[0][6]
				MinN = strings.Replace(MinN, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 7 {
				MinNStar = SearchQueryFromNOS[0][7]
				MinNStar = strings.Replace(MinNStar, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 8 {
				NDelta = SearchQueryFromNOS[0][8]
				NDelta = strings.Replace(NDelta, ",", "", -1)
			}
		}
	}

	SessionID := ctx.UserValue("SessionID")
	if SessionID != nil {
		temp := SessionID.(string)
		temp = temp[1 : len(temp)-1]
		SessionID = temp
	}
	NoOneSucceededBoolean := false
	var CookieErrorString1 string
	SourceSearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearchSheetName + "!A4:G5000")
	sourceIterator := 0
	loc, _ := time.LoadLocation("Europe/Rome")
	currentTime := time.Now().In(loc)
	Time := currentTime.Format("2006-01-02")
	for sourceIterator < len(SourceSearchQueryFromNOS) {
		if len(SourceSearchQueryFromNOS[sourceIterator]) < 2 {
			return
		}
		fmt.Println(MinFollower)
		fmt.Println(MaxFollower)
		fmt.Println(MinN)
		fmt.Println(MinNStar)
		fmt.Println(NDelta)
		fmt.Println(SessionID)
		userName := SourceSearchQueryFromNOS[sourceIterator][0]
		if userName == "" {
			sugar.Infof("queryString for search is nil ")
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetStatusCode(200)
			ctx.SetBody([]byte("Failed! Unable to Find USERNAME shared in URL"))
			sugar.Infof("calling ig reprts failure due to username!")
			return
		}
		LastFetchedFollowerCount := SourceSearchQueryFromNOS[sourceIterator][1]
		if LastFetchedFollowerCount == "" {
			LastFetchedFollowerCount = "10"
		}
		fmt.Println(userName)
		fmt.Println(LastFetchedFollowerCount)
		FollowersList, _, LatestFollowerCount := ig.GetNewFollowers(userName, LastFetchedFollowerCount, SessionID.(string))
		var nosLatestFollowerCountRows []interface{}
		if LatestFollowerCount != 0 {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName, LatestFollowerCount)
		} else {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName)
		}
		nosLatestFollowerCountFinalValues = append(nosLatestFollowerCountFinalValues, nosLatestFollowerCountRows)
		fmt.Println(FollowersList)
		SearchQuery := make(map[string]int)
		if MinFollower != "" {
			temp := MinFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinFollower"] = tempInt
			}
		}
		if MaxFollower != "" {
			temp := MaxFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MaxFollower"] = tempInt
			}
		}
		if MinN != "" {
			temp := MinN
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinN"] = tempInt
			}
		}
		if MinNStar != "" {
			temp := MinNStar
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinNStar"] = tempInt
			}
		}
		NDeltaFloat := 0.0
		if NDelta != "" {
			temp := NDelta
			tempFloat, e := strconv.ParseFloat(temp, 2)
			if e == nil {
				NDeltaFloat = tempFloat
			}
		}
		var reportValues [][]interface{}
		reportValues, NoOneSucceededBoolean, CookieErrorString1 = ig.GetIGReportNew(FollowersList, SearchQuery, SessionID.(string), NDeltaFloat)
		fmt.Println(reportValues)
		i := 0
		for i < len(reportValues) {
			var searchRow []interface{}
			var dashboardRow []interface{}
			if (len(reportValues[i])) == 7 {
				dashboardRow = append(dashboardRow, Time, "#1", SourceSearchQueryFromNOS[sourceIterator][2], SourceSearchQueryFromNOS[sourceIterator][3], SourceSearchQueryFromNOS[sourceIterator][4], SourceSearchQueryFromNOS[sourceIterator][5], userName, reportValues[i][0], reportValues[i][1], reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6])
				nosDashboardFinalValues = append(nosDashboardFinalValues, dashboardRow)
				searchRow = append(searchRow, Time, reportValues[i][0], userName, reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6])
				nosSearchFinalValues = append(nosSearchFinalValues, searchRow)
			}
			i++
		}
		sourceIterator++
	}
	fmt.Println("*********")
	fmt.Println(nosDashboardFinalValues)
	fmt.Println("#########")
	fmt.Println(nosSearchFinalValues)
	if len(nosDashboardFinalValues) > 0 {
		googleSheets.BatchAppend(configs.Configurations.NOSDashboardSheetName, nosDashboardFinalValues)
		existingRows := googleSheets.BatchGet(configs.Configurations.NOSSearchSheetName + "!C4:I5000")
		StartingRow := len(existingRows) + 3 + 1
		googleSheets.BatchWrite(configs.Configurations.NOSSearchSheetName+"!C"+strconv.Itoa(StartingRow)+":I5000", nosSearchFinalValues)
		googleSheets.BatchWrite(configs.Configurations.NOSSearchSheetName+"!A4:B5000", nosLatestFollowerCountFinalValues)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Google Sheet Updated" + " -- " + CookieErrorString1))
		sugar.Infof("calling ig research reports success!")
	} else if NoOneSucceededBoolean {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Noone passed the filter search query"))
		sugar.Infof("calling ig research reports success!" + " -- " + CookieErrorString1)
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, not able to fetch data"))
		sugar.Infof("calling ig research reports failure!" + " -- " + CookieErrorString1)
	}
}

func handleNOSSearchSetupLatest(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a NOS Search request to Google Sheets!")
	SearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch2SheetName + "!A2:N2")
	fmt.Println(SearchQueryFromNOS)
	var nosSearchFinalValues [][]interface{}
	var nosDashboardFinalValues [][]interface{}
	var nosLatestFollowerCountFinalValues [][]interface{}
	var MinFollower string
	var MaxFollower string
	var MinN string
	var MinNStar string
	var NDelta string

	if len(SearchQueryFromNOS) == 1 {
		if len(SearchQueryFromNOS[0]) > 9 {
			MinFollower = SearchQueryFromNOS[0][8]
			MaxFollower = SearchQueryFromNOS[0][10]
			MinFollower = strings.Replace(MinFollower, ",", "", -1)
			MaxFollower = strings.Replace(MaxFollower, ",", "", -1)
			if len(SearchQueryFromNOS[0]) > 11 {
				MinN = SearchQueryFromNOS[0][11]
				MinN = strings.Replace(MinN, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 12 {
				MinNStar = SearchQueryFromNOS[0][12]
				MinNStar = strings.Replace(MinNStar, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 13 {
				NDelta = SearchQueryFromNOS[0][13]
				NDelta = strings.Replace(NDelta, ",", "", -1)
			}
		}
	}

	SessionID := ctx.UserValue("SessionID")
	if SessionID != nil {
		temp := SessionID.(string)
		temp = temp[1 : len(temp)-1]
		SessionID = temp
		CookieFinder := googleSheets.BatchGet(configs.Configurations.CookieFinderSheet + "!A4:B4")
		if len(CookieFinder) == 1 {
			if len(CookieFinder[0]) == 2 {
				SessionID = CookieFinder[0][1]
				fmt.Print("Received Session ids from Sheet: ")
				fmt.Println(SessionID.(string))
			}
		}
	}
	NoOneSucceededBoolean := false
	var CookieErrorString1 string
	SourceSearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch2SheetName + "!A4:G5000")
	sourceIterator := 0
	loc, _ := time.LoadLocation("Europe/Rome")
	currentTime := time.Now().In(loc)
	Time := currentTime.Format("2006-01-02")
	for sourceIterator < len(SourceSearchQueryFromNOS) {
		if len(SourceSearchQueryFromNOS[sourceIterator]) < 2 {
			sourceIterator++
			continue
		}
		fmt.Println(MinFollower)
		fmt.Println(MaxFollower)
		fmt.Println(MinN)
		fmt.Println(MinNStar)
		fmt.Println(NDelta)
		fmt.Println(SessionID)
		userName := SourceSearchQueryFromNOS[sourceIterator][0]
		if userName == "" {
			sugar.Infof("queryString for search is nil ")
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetStatusCode(200)
			ctx.SetBody([]byte("Failed! Unable to Find USERNAME shared in URL"))
			sugar.Infof("calling ig reprts failure due to username!")
			sourceIterator++
			continue
		}
		LastFetchedFollowerCount := SourceSearchQueryFromNOS[sourceIterator][1]
		if LastFetchedFollowerCount == "" {
			LastFetchedFollowerCount = "10"
		}
		fmt.Println(userName)
		fmt.Println(LastFetchedFollowerCount)
		FollowersList, _, LatestFollowerCount := ig.GetNewFollowers(userName, LastFetchedFollowerCount, SessionID.(string))
		var nosLatestFollowerCountRows []interface{}
		if LatestFollowerCount != 0 {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName, LatestFollowerCount)
		} else {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName)
		}
		nosLatestFollowerCountFinalValues = append(nosLatestFollowerCountFinalValues, nosLatestFollowerCountRows)
		fmt.Println(FollowersList)
		SearchQuery := make(map[string]int)
		if MinFollower != "" {
			temp := MinFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinFollower"] = tempInt
			}
		}
		if MaxFollower != "" {
			temp := MaxFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MaxFollower"] = tempInt
			}
		}
		if MinN != "" {
			temp := MinN
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinN"] = tempInt
			}
		}
		if MinNStar != "" {
			temp := MinNStar
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinNStar"] = tempInt
			}
		}
		NDeltaFloat := 0.0
		if NDelta != "" {
			temp := NDelta
			tempFloat, e := strconv.ParseFloat(temp, 2)
			if e == nil {
				NDeltaFloat = tempFloat
			}
		}
		var reportValues [][]interface{}
		reportValues, NoOneSucceededBoolean, CookieErrorString1 = ig.GetIGReportNew(FollowersList, SearchQuery, SessionID.(string), NDeltaFloat)
		fmt.Println(reportValues)
		i := 0
		for i < len(reportValues) {
			var searchRow []interface{}
			var dashboardRow []interface{}
			if (len(reportValues[i])) > 5 {
				dashboardRow = append(dashboardRow, Time, "#2", SourceSearchQueryFromNOS[sourceIterator][2], SourceSearchQueryFromNOS[sourceIterator][3], SourceSearchQueryFromNOS[sourceIterator][4], SourceSearchQueryFromNOS[sourceIterator][5], SourceSearchQueryFromNOS[sourceIterator][6], reportValues[i][0], reportValues[i][1], reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6], reportValues[i][7], reportValues[i][8])
				nosDashboardFinalValues = append(nosDashboardFinalValues, dashboardRow)
				searchRow = append(searchRow, Time, reportValues[i][0], userName, reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6])
				nosSearchFinalValues = append(nosSearchFinalValues, searchRow)
			}
			i++
		}
		sourceIterator++
	}
	fmt.Println("*********")
	fmt.Println(nosDashboardFinalValues)
	fmt.Println("#########")
	fmt.Println(nosSearchFinalValues)
	if len(nosDashboardFinalValues) > 0 {
		googleSheets.BatchAppend(configs.Configurations.NOSDashboardSheetName, nosDashboardFinalValues)
		existingRows := googleSheets.BatchGet(configs.Configurations.NOSSearch2SheetName + "!H4:N5000")
		StartingRow := len(existingRows) + 3 + 1
		googleSheets.BatchWrite(configs.Configurations.NOSSearch2SheetName+"!H"+strconv.Itoa(StartingRow)+":N5000", nosSearchFinalValues)
		googleSheets.BatchWrite(configs.Configurations.NOSSearch2SheetName+"!A4:B5000", nosLatestFollowerCountFinalValues)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Google Sheet Updated" + " -- " + CookieErrorString1))
		sugar.Infof("calling ig research reports success!")
	} else if NoOneSucceededBoolean {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Noone passed the filter search query"))
		sugar.Infof("calling ig research reports success!" + " -- " + CookieErrorString1)
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, not able to fetch data"))
		sugar.Infof("calling ig research reports failure!" + " -- " + CookieErrorString1)
	}
}

func handleNOSSearchSetup4(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a NOS Search request to Google Sheets!")
	SearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch4SheetName + "!A2:N2")
	fmt.Println(SearchQueryFromNOS)
	var nosSearchFinalValues [][]interface{}
	var nosDashboardFinalValues [][]interface{}
	var nosLatestFollowerCountFinalValues [][]interface{}
	var MinFollower string
	var MaxFollower string
	var MinN string
	var MinNStar string
	var NDelta string

	if len(SearchQueryFromNOS) == 1 {
		if len(SearchQueryFromNOS[0]) > 9 {
			MinFollower = SearchQueryFromNOS[0][8]
			MaxFollower = SearchQueryFromNOS[0][10]
			MinFollower = strings.Replace(MinFollower, ",", "", -1)
			MaxFollower = strings.Replace(MaxFollower, ",", "", -1)
			if len(SearchQueryFromNOS[0]) > 11 {
				MinN = SearchQueryFromNOS[0][11]
				MinN = strings.Replace(MinN, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 12 {
				MinNStar = SearchQueryFromNOS[0][12]
				MinNStar = strings.Replace(MinNStar, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 13 {
				NDelta = SearchQueryFromNOS[0][13]
				NDelta = strings.Replace(NDelta, ",", "", -1)
			}
		}
	}

	SessionID := ctx.UserValue("SessionID")
	if SessionID != nil {
		temp := SessionID.(string)
		temp = temp[1 : len(temp)-1]
		SessionID = temp
		CookieFinder := googleSheets.BatchGet(configs.Configurations.CookieFinderSheet + "!A6:B6")
		if len(CookieFinder) == 1 {
			if len(CookieFinder[0]) == 2 {
				SessionID = CookieFinder[0][1]
				fmt.Print("Received Session ids from Sheet: ")
				fmt.Println(SessionID.(string))
			}
		}
	}
	NoOneSucceededBoolean := false
	var CookieErrorString1 string
	SourceSearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch4SheetName + "!A4:G5000")
	sourceIterator := 0
	loc, _ := time.LoadLocation("Europe/Rome")
	currentTime := time.Now().In(loc)
	Time := currentTime.Format("2006-01-02")
	for sourceIterator < len(SourceSearchQueryFromNOS) {
		if len(SourceSearchQueryFromNOS[sourceIterator]) < 2 {
			sourceIterator++
			continue
		}
		fmt.Println(MinFollower)
		fmt.Println(MaxFollower)
		fmt.Println(MinN)
		fmt.Println(MinNStar)
		fmt.Println(NDelta)
		fmt.Println(SessionID)
		userName := SourceSearchQueryFromNOS[sourceIterator][0]
		if userName == "" {
			sugar.Infof("queryString for search is nil ")
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetStatusCode(200)
			ctx.SetBody([]byte("Failed! Unable to Find USERNAME shared in URL"))
			sugar.Infof("calling ig reprts failure due to username!")
			sourceIterator++
			continue
		}
		LastFetchedFollowerCount := SourceSearchQueryFromNOS[sourceIterator][1]
		if LastFetchedFollowerCount == "" {
			LastFetchedFollowerCount = "10"
		}
		fmt.Println(userName)
		fmt.Println(LastFetchedFollowerCount)
		FollowersList, _, LatestFollowerCount := ig.GetNewFollowers(userName, LastFetchedFollowerCount, SessionID.(string))
		var nosLatestFollowerCountRows []interface{}
		if LatestFollowerCount != 0 {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName, LatestFollowerCount)
		} else {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName)
		}
		nosLatestFollowerCountFinalValues = append(nosLatestFollowerCountFinalValues, nosLatestFollowerCountRows)
		fmt.Println(FollowersList)
		SearchQuery := make(map[string]int)
		if MinFollower != "" {
			temp := MinFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinFollower"] = tempInt
			}
		}
		if MaxFollower != "" {
			temp := MaxFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MaxFollower"] = tempInt
			}
		}
		if MinN != "" {
			temp := MinN
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinN"] = tempInt
			}
		}
		if MinNStar != "" {
			temp := MinNStar
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinNStar"] = tempInt
			}
		}
		NDeltaFloat := 0.0
		if NDelta != "" {
			temp := NDelta
			tempFloat, e := strconv.ParseFloat(temp, 2)
			if e == nil {
				NDeltaFloat = tempFloat
			}
		}
		var reportValues [][]interface{}
		reportValues, NoOneSucceededBoolean, CookieErrorString1 = ig.GetIGReportNew(FollowersList, SearchQuery, SessionID.(string), NDeltaFloat)
		fmt.Println(reportValues)
		i := 0
		for i < len(reportValues) {
			var searchRow []interface{}
			var dashboardRow []interface{}
			if (len(reportValues[i])) > 5 {
				dashboardRow = append(dashboardRow, Time, "#4", SourceSearchQueryFromNOS[sourceIterator][2], SourceSearchQueryFromNOS[sourceIterator][3], SourceSearchQueryFromNOS[sourceIterator][4], SourceSearchQueryFromNOS[sourceIterator][5], SourceSearchQueryFromNOS[sourceIterator][6], reportValues[i][0], reportValues[i][1], reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6], reportValues[i][7], reportValues[i][8])
				nosDashboardFinalValues = append(nosDashboardFinalValues, dashboardRow)
				searchRow = append(searchRow, Time, reportValues[i][0], userName, reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6])
				nosSearchFinalValues = append(nosSearchFinalValues, searchRow)
			}
			i++
		}
		sourceIterator++
	}
	fmt.Println("*********")
	fmt.Println(nosDashboardFinalValues)
	fmt.Println("#########")
	fmt.Println(nosSearchFinalValues)
	if len(nosDashboardFinalValues) > 0 {
		googleSheets.BatchAppend(configs.Configurations.NOSDashboardSheetName, nosDashboardFinalValues)
		existingRows := googleSheets.BatchGet(configs.Configurations.NOSSearch4SheetName + "!H4:N5000")
		StartingRow := len(existingRows) + 3 + 1
		googleSheets.BatchWrite(configs.Configurations.NOSSearch4SheetName+"!H"+strconv.Itoa(StartingRow)+":N5000", nosSearchFinalValues)
		googleSheets.BatchWrite(configs.Configurations.NOSSearch4SheetName+"!A4:B5000", nosLatestFollowerCountFinalValues)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Google Sheet Updated" + " -- " + CookieErrorString1))
		sugar.Infof("calling ig research reports success!")
	} else if NoOneSucceededBoolean {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Noone passed the filter search query"))
		sugar.Infof("calling ig research reports success!" + " -- " + CookieErrorString1)
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, not able to fetch data"))
		sugar.Infof("calling ig research reports failure!" + " -- " + CookieErrorString1)
	}
}

func handleNOSSearchSetup5(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a NOS Search request to Google Sheets!")
	SearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch5SheetName + "!A2:N2")
	fmt.Println(SearchQueryFromNOS)
	var nosSearchFinalValues [][]interface{}
	var nosDashboardFinalValues [][]interface{}
	var nosLatestFollowerCountFinalValues [][]interface{}
	var MinFollower string
	var MaxFollower string
	var MinN string
	var MinNStar string
	var NDelta string

	if len(SearchQueryFromNOS) == 1 {
		if len(SearchQueryFromNOS[0]) > 9 {
			MinFollower = SearchQueryFromNOS[0][8]
			MaxFollower = SearchQueryFromNOS[0][10]
			MinFollower = strings.Replace(MinFollower, ",", "", -1)
			MaxFollower = strings.Replace(MaxFollower, ",", "", -1)
			if len(SearchQueryFromNOS[0]) > 11 {
				MinN = SearchQueryFromNOS[0][11]
				MinN = strings.Replace(MinN, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 12 {
				MinNStar = SearchQueryFromNOS[0][12]
				MinNStar = strings.Replace(MinNStar, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 13 {
				NDelta = SearchQueryFromNOS[0][13]
				NDelta = strings.Replace(NDelta, ",", "", -1)
			}
		}
	}

	SessionID := ctx.UserValue("SessionID")
	if SessionID != nil {
		temp := SessionID.(string)
		temp = temp[1 : len(temp)-1]
		SessionID = temp
		CookieFinder := googleSheets.BatchGet(configs.Configurations.CookieFinderSheet + "!A7:B7")
		if len(CookieFinder) == 1 {
			if len(CookieFinder[0]) == 2 {
				SessionID = CookieFinder[0][1]
				fmt.Print("Received Session ids from Sheet: ")
				fmt.Println(SessionID.(string))
			}
		}
	}
	NoOneSucceededBoolean := false
	var CookieErrorString1 string
	SourceSearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch5SheetName + "!A4:G5000")
	sourceIterator := 0
	loc, _ := time.LoadLocation("Europe/Rome")
	currentTime := time.Now().In(loc)
	Time := currentTime.Format("2006-01-02")
	for sourceIterator < len(SourceSearchQueryFromNOS) {
		if len(SourceSearchQueryFromNOS[sourceIterator]) < 2 {
			sourceIterator++
			continue
		}
		fmt.Println(MinFollower)
		fmt.Println(MaxFollower)
		fmt.Println(MinN)
		fmt.Println(MinNStar)
		fmt.Println(NDelta)
		fmt.Println(SessionID)
		userName := SourceSearchQueryFromNOS[sourceIterator][0]
		if userName == "" {
			sugar.Infof("queryString for search is nil ")
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetStatusCode(200)
			ctx.SetBody([]byte("Failed! Unable to Find USERNAME shared in URL"))
			sugar.Infof("calling ig reprts failure due to username!")
			sourceIterator++
			continue
		}
		LastFetchedFollowerCount := SourceSearchQueryFromNOS[sourceIterator][1]
		if LastFetchedFollowerCount == "" {
			LastFetchedFollowerCount = "10"
		}
		fmt.Println(userName)
		fmt.Println(LastFetchedFollowerCount)
		FollowersList, _, LatestFollowerCount := ig.GetNewFollowers(userName, LastFetchedFollowerCount, SessionID.(string))
		var nosLatestFollowerCountRows []interface{}
		if LatestFollowerCount != 0 {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName, LatestFollowerCount)
		} else {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName)
		}
		nosLatestFollowerCountFinalValues = append(nosLatestFollowerCountFinalValues, nosLatestFollowerCountRows)
		fmt.Println(FollowersList)
		SearchQuery := make(map[string]int)
		if MinFollower != "" {
			temp := MinFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinFollower"] = tempInt
			}
		}
		if MaxFollower != "" {
			temp := MaxFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MaxFollower"] = tempInt
			}
		}
		if MinN != "" {
			temp := MinN
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinN"] = tempInt
			}
		}
		if MinNStar != "" {
			temp := MinNStar
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinNStar"] = tempInt
			}
		}
		NDeltaFloat := 0.0
		if NDelta != "" {
			temp := NDelta
			tempFloat, e := strconv.ParseFloat(temp, 2)
			if e == nil {
				NDeltaFloat = tempFloat
			}
		}
		var reportValues [][]interface{}
		reportValues, NoOneSucceededBoolean, CookieErrorString1 = ig.GetIGReportNew(FollowersList, SearchQuery, SessionID.(string), NDeltaFloat)
		fmt.Println(reportValues)
		i := 0
		for i < len(reportValues) {
			var searchRow []interface{}
			var dashboardRow []interface{}
			if (len(reportValues[i])) > 5 {
				dashboardRow = append(dashboardRow, Time, "#5", SourceSearchQueryFromNOS[sourceIterator][2], SourceSearchQueryFromNOS[sourceIterator][3], SourceSearchQueryFromNOS[sourceIterator][4], SourceSearchQueryFromNOS[sourceIterator][5], SourceSearchQueryFromNOS[sourceIterator][6], reportValues[i][0], reportValues[i][1], reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6], reportValues[i][7], reportValues[i][8])
				nosDashboardFinalValues = append(nosDashboardFinalValues, dashboardRow)
				searchRow = append(searchRow, Time, reportValues[i][0], userName, reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6])
				nosSearchFinalValues = append(nosSearchFinalValues, searchRow)
			}
			i++
		}
		sourceIterator++
	}
	fmt.Println("*********")
	fmt.Println(nosDashboardFinalValues)
	fmt.Println("#########")
	fmt.Println(nosSearchFinalValues)
	if len(nosDashboardFinalValues) > 0 {
		googleSheets.BatchAppend(configs.Configurations.NOSDashboardSheetName, nosDashboardFinalValues)
		existingRows := googleSheets.BatchGet(configs.Configurations.NOSSearch5SheetName + "!H4:N5000")
		StartingRow := len(existingRows) + 3 + 1
		googleSheets.BatchWrite(configs.Configurations.NOSSearch5SheetName+"!H"+strconv.Itoa(StartingRow)+":N5000", nosSearchFinalValues)
		googleSheets.BatchWrite(configs.Configurations.NOSSearch5SheetName+"!A4:B5000", nosLatestFollowerCountFinalValues)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Google Sheet Updated" + " -- " + CookieErrorString1))
		sugar.Infof("calling ig research reports success!")
	} else if NoOneSucceededBoolean {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Noone passed the filter search query"))
		sugar.Infof("calling ig research reports success!" + " -- " + CookieErrorString1)
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, not able to fetch data"))
		sugar.Infof("calling ig research reports failure!" + " -- " + CookieErrorString1)
	}
}

func handleNOSSearchSetup6(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a NOS Search request to Google Sheets!")
	SearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch6SheetName + "!A2:N2")
	fmt.Println(SearchQueryFromNOS)
	var nosSearchFinalValues [][]interface{}
	var nosDashboardFinalValues [][]interface{}
	var nosLatestFollowerCountFinalValues [][]interface{}
	var MinFollower string
	var MaxFollower string
	var MinN string
	var MinNStar string
	var NDelta string

	if len(SearchQueryFromNOS) == 1 {
		if len(SearchQueryFromNOS[0]) > 9 {
			MinFollower = SearchQueryFromNOS[0][8]
			MaxFollower = SearchQueryFromNOS[0][10]
			MinFollower = strings.Replace(MinFollower, ",", "", -1)
			MaxFollower = strings.Replace(MaxFollower, ",", "", -1)
			if len(SearchQueryFromNOS[0]) > 11 {
				MinN = SearchQueryFromNOS[0][11]
				MinN = strings.Replace(MinN, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 12 {
				MinNStar = SearchQueryFromNOS[0][12]
				MinNStar = strings.Replace(MinNStar, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 13 {
				NDelta = SearchQueryFromNOS[0][13]
				NDelta = strings.Replace(NDelta, ",", "", -1)
			}
		}
	}

	SessionID := ctx.UserValue("SessionID")
	if SessionID != nil {
		temp := SessionID.(string)
		temp = temp[1 : len(temp)-1]
		SessionID = temp
		CookieFinder := googleSheets.BatchGet(configs.Configurations.CookieFinderSheet + "!A8:B8")
		if len(CookieFinder) == 1 {
			if len(CookieFinder[0]) == 2 {
				SessionID = CookieFinder[0][1]
				fmt.Print("Received Session ids from Sheet: ")
				fmt.Println(SessionID.(string))
			}
		}
	}
	NoOneSucceededBoolean := false
	var CookieErrorString1 string
	SourceSearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch6SheetName + "!A4:G5000")
	sourceIterator := 0
	loc, _ := time.LoadLocation("Europe/Rome")
	currentTime := time.Now().In(loc)
	Time := currentTime.Format("2006-01-02")
	for sourceIterator < len(SourceSearchQueryFromNOS) {
		if len(SourceSearchQueryFromNOS[sourceIterator]) < 2 {
			sourceIterator++
			continue
		}
		fmt.Println(MinFollower)
		fmt.Println(MaxFollower)
		fmt.Println(MinN)
		fmt.Println(MinNStar)
		fmt.Println(NDelta)
		fmt.Println(SessionID)
		userName := SourceSearchQueryFromNOS[sourceIterator][0]
		if userName == "" {
			sugar.Infof("queryString for search is nil ")
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetStatusCode(200)
			ctx.SetBody([]byte("Failed! Unable to Find USERNAME shared in URL"))
			sugar.Infof("calling ig reprts failure due to username!")
			sourceIterator++
			continue
		}
		LastFetchedFollowerCount := SourceSearchQueryFromNOS[sourceIterator][1]
		if LastFetchedFollowerCount == "" {
			LastFetchedFollowerCount = "10"
		}
		fmt.Println(userName)
		fmt.Println(LastFetchedFollowerCount)
		FollowersList, _, LatestFollowerCount := ig.GetNewFollowers(userName, LastFetchedFollowerCount, SessionID.(string))
		var nosLatestFollowerCountRows []interface{}
		if LatestFollowerCount != 0 {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName, LatestFollowerCount)
		} else {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName)
		}
		nosLatestFollowerCountFinalValues = append(nosLatestFollowerCountFinalValues, nosLatestFollowerCountRows)
		fmt.Println(FollowersList)
		SearchQuery := make(map[string]int)
		if MinFollower != "" {
			temp := MinFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinFollower"] = tempInt
			}
		}
		if MaxFollower != "" {
			temp := MaxFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MaxFollower"] = tempInt
			}
		}
		if MinN != "" {
			temp := MinN
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinN"] = tempInt
			}
		}
		if MinNStar != "" {
			temp := MinNStar
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinNStar"] = tempInt
			}
		}
		NDeltaFloat := 0.0
		if NDelta != "" {
			temp := NDelta
			tempFloat, e := strconv.ParseFloat(temp, 2)
			if e == nil {
				NDeltaFloat = tempFloat
			}
		}
		var reportValues [][]interface{}
		reportValues, NoOneSucceededBoolean, CookieErrorString1 = ig.GetIGReportNew(FollowersList, SearchQuery, SessionID.(string), NDeltaFloat)
		fmt.Println(reportValues)
		i := 0
		for i < len(reportValues) {
			var searchRow []interface{}
			var dashboardRow []interface{}
			if (len(reportValues[i])) > 5 {
				dashboardRow = append(dashboardRow, Time, "#6", SourceSearchQueryFromNOS[sourceIterator][2], SourceSearchQueryFromNOS[sourceIterator][3], SourceSearchQueryFromNOS[sourceIterator][4], SourceSearchQueryFromNOS[sourceIterator][5], SourceSearchQueryFromNOS[sourceIterator][6], reportValues[i][0], reportValues[i][1], reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6], reportValues[i][7], reportValues[i][8])
				nosDashboardFinalValues = append(nosDashboardFinalValues, dashboardRow)
				searchRow = append(searchRow, Time, reportValues[i][0], userName, reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6])
				nosSearchFinalValues = append(nosSearchFinalValues, searchRow)
			}
			i++
		}
		sourceIterator++
	}
	fmt.Println("*********")
	fmt.Println(nosDashboardFinalValues)
	fmt.Println("#########")
	fmt.Println(nosSearchFinalValues)
	if len(nosDashboardFinalValues) > 0 {
		googleSheets.BatchAppend(configs.Configurations.NOSDashboardSheetName, nosDashboardFinalValues)
		existingRows := googleSheets.BatchGet(configs.Configurations.NOSSearch6SheetName + "!H4:N5000")
		StartingRow := len(existingRows) + 3 + 1
		googleSheets.BatchWrite(configs.Configurations.NOSSearch6SheetName+"!H"+strconv.Itoa(StartingRow)+":N5000", nosSearchFinalValues)
		googleSheets.BatchWrite(configs.Configurations.NOSSearch6SheetName+"!A4:B5000", nosLatestFollowerCountFinalValues)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Google Sheet Updated" + " -- " + CookieErrorString1))
		sugar.Infof("calling ig research reports success!")
	} else if NoOneSucceededBoolean {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Noone passed the filter search query"))
		sugar.Infof("calling ig research reports success!" + " -- " + CookieErrorString1)
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, not able to fetch data"))
		sugar.Infof("calling ig research reports failure!" + " -- " + CookieErrorString1)
	}
}

func handleNOSSearchSetup3(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a NOS Search request to Google Sheets!")
	SearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch3SheetName + "!A2:N2")
	fmt.Println(SearchQueryFromNOS)
	var nosSearchFinalValues [][]interface{}
	var nosDashboardFinalValues [][]interface{}
	var nosLatestFollowerCountFinalValues [][]interface{}
	var MinFollower string
	var MaxFollower string
	var MinN string
	var MinNStar string
	var NDelta string

	if len(SearchQueryFromNOS) == 1 {
		if len(SearchQueryFromNOS[0]) > 9 {
			MinFollower = SearchQueryFromNOS[0][8]
			MaxFollower = SearchQueryFromNOS[0][10]
			MinFollower = strings.Replace(MinFollower, ",", "", -1)
			MaxFollower = strings.Replace(MaxFollower, ",", "", -1)
			if len(SearchQueryFromNOS[0]) > 11 {
				MinN = SearchQueryFromNOS[0][11]
				MinN = strings.Replace(MinN, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 12 {
				MinNStar = SearchQueryFromNOS[0][12]
				MinNStar = strings.Replace(MinNStar, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 13 {
				NDelta = SearchQueryFromNOS[0][13]
				NDelta = strings.Replace(NDelta, ",", "", -1)
			}
		}
	}

	SessionID := ctx.UserValue("SessionID")
	if SessionID != nil {
		temp := SessionID.(string)
		temp = temp[1 : len(temp)-1]
		SessionID = temp
		CookieFinder := googleSheets.BatchGet(configs.Configurations.CookieFinderSheet + "!A5:B5")
		if len(CookieFinder) == 1 {
			if len(CookieFinder[0]) == 2 {
				SessionID = CookieFinder[0][1]
				fmt.Print("Received Session ids from Sheet: ")
				fmt.Println(SessionID.(string))
			}
		}
	}
	NoOneSucceededBoolean := false
	var CookieErrorString1 string
	SourceSearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch3SheetName + "!A4:G5000")
	sourceIterator := 0
	loc, _ := time.LoadLocation("Europe/Rome")
	currentTime := time.Now().In(loc)
	Time := currentTime.Format("2006-01-02")
	for sourceIterator < len(SourceSearchQueryFromNOS) {
		if len(SourceSearchQueryFromNOS[sourceIterator]) < 2 {
			sourceIterator++
			continue
		}
		fmt.Println(MinFollower)
		fmt.Println(MaxFollower)
		fmt.Println(MinN)
		fmt.Println(MinNStar)
		fmt.Println(NDelta)
		fmt.Println(SessionID)
		userName := SourceSearchQueryFromNOS[sourceIterator][0]
		if userName == "" {
			sugar.Infof("queryString for search is nil ")
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetStatusCode(200)
			ctx.SetBody([]byte("Failed! Unable to Find USERNAME shared in URL"))
			sugar.Infof("calling ig reprts failure due to username!")
			sourceIterator++
			continue
		}
		LastFetchedFollowerCount := SourceSearchQueryFromNOS[sourceIterator][1]
		if LastFetchedFollowerCount == "" {
			LastFetchedFollowerCount = "10"
		}
		fmt.Println(userName)
		fmt.Println(LastFetchedFollowerCount)
		FollowersList, _, LatestFollowerCount := ig.GetNewFollowers(userName, LastFetchedFollowerCount, SessionID.(string))
		var nosLatestFollowerCountRows []interface{}
		if LatestFollowerCount != 0 {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName, LatestFollowerCount)
		} else {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName)
		}
		nosLatestFollowerCountFinalValues = append(nosLatestFollowerCountFinalValues, nosLatestFollowerCountRows)
		fmt.Println(FollowersList)
		SearchQuery := make(map[string]int)
		if MinFollower != "" {
			temp := MinFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinFollower"] = tempInt
			}
		}
		if MaxFollower != "" {
			temp := MaxFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MaxFollower"] = tempInt
			}
		}
		if MinN != "" {
			temp := MinN
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinN"] = tempInt
			}
		}
		if MinNStar != "" {
			temp := MinNStar
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinNStar"] = tempInt
			}
		}
		NDeltaFloat := 0.0
		if NDelta != "" {
			temp := NDelta
			tempFloat, e := strconv.ParseFloat(temp, 2)
			if e == nil {
				NDeltaFloat = tempFloat
			}
		}
		var reportValues [][]interface{}
		reportValues, NoOneSucceededBoolean, CookieErrorString1 = ig.GetIGReportNew(FollowersList, SearchQuery, SessionID.(string), NDeltaFloat)
		fmt.Println(reportValues)
		i := 0
		for i < len(reportValues) {
			var searchRow []interface{}
			var dashboardRow []interface{}
			if (len(reportValues[i])) > 5 {
				dashboardRow = append(dashboardRow, Time, "#3", SourceSearchQueryFromNOS[sourceIterator][2], SourceSearchQueryFromNOS[sourceIterator][3], SourceSearchQueryFromNOS[sourceIterator][4], SourceSearchQueryFromNOS[sourceIterator][5], SourceSearchQueryFromNOS[sourceIterator][6], reportValues[i][0], reportValues[i][1], reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6], reportValues[i][7], reportValues[i][8])
				nosDashboardFinalValues = append(nosDashboardFinalValues, dashboardRow)
				searchRow = append(searchRow, Time, reportValues[i][0], userName, reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6])
				nosSearchFinalValues = append(nosSearchFinalValues, searchRow)
			}
			i++
		}
		sourceIterator++
	}
	fmt.Println("*********")
	fmt.Println(nosDashboardFinalValues)
	fmt.Println("#########")
	fmt.Println(nosSearchFinalValues)
	if len(nosDashboardFinalValues) > 0 {
		googleSheets.BatchAppend(configs.Configurations.NOSDashboardSheetName, nosDashboardFinalValues)
		existingRows := googleSheets.BatchGet(configs.Configurations.NOSSearch3SheetName + "!H4:N5000")
		StartingRow := len(existingRows) + 3 + 1
		googleSheets.BatchWrite(configs.Configurations.NOSSearch3SheetName+"!H"+strconv.Itoa(StartingRow)+":N5000", nosSearchFinalValues)
		googleSheets.BatchWrite(configs.Configurations.NOSSearch3SheetName+"!A4:B5000", nosLatestFollowerCountFinalValues)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Google Sheet Updated" + " -- " + CookieErrorString1))
		sugar.Infof("calling ig research reports success!")
	} else if NoOneSucceededBoolean {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Noone passed the filter search query"))
		sugar.Infof("calling ig research reports success!" + " -- " + CookieErrorString1)
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, not able to fetch data"))
		sugar.Infof("calling ig research reports failure!" + " -- " + CookieErrorString1)
	}
}

func handleNOSSearchSetup1(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a NOS Search 1 request to Google Sheets!")
	SearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearchSheetName + "!A2:N2")
	fmt.Println(SearchQueryFromNOS)
	var nosSearchFinalValues [][]interface{}
	var nosDashboardFinalValues [][]interface{}
	var nosLatestFollowerCountFinalValues [][]interface{}
	var MinFollower string
	var MaxFollower string
	var MinN string
	var MinNStar string
	var NDelta string

	if len(SearchQueryFromNOS) == 1 {
		if len(SearchQueryFromNOS[0]) > 9 {
			MinFollower = SearchQueryFromNOS[0][8]
			MaxFollower = SearchQueryFromNOS[0][10]
			MinFollower = strings.Replace(MinFollower, ",", "", -1)
			MaxFollower = strings.Replace(MaxFollower, ",", "", -1)
			if len(SearchQueryFromNOS[0]) > 11 {
				MinN = SearchQueryFromNOS[0][11]
				MinN = strings.Replace(MinN, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 12 {
				MinNStar = SearchQueryFromNOS[0][12]
				MinNStar = strings.Replace(MinNStar, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 13 {
				NDelta = SearchQueryFromNOS[0][13]
				NDelta = strings.Replace(NDelta, ",", "", -1)
			}
		}
	}

	SessionID := ctx.UserValue("SessionID")
	if SessionID != nil {
		temp := SessionID.(string)
		temp = temp[1 : len(temp)-1]
		SessionID = temp
		CookieFinder := googleSheets.BatchGet(configs.Configurations.CookieFinderSheet + "!A3:B3")
		if len(CookieFinder) == 1 {
			if len(CookieFinder[0]) == 2 {
				SessionID = CookieFinder[0][1]
				fmt.Print("Received Session ids from Sheet: ")
				fmt.Println(SessionID.(string))
			}
		}
	}
	// NoOneSucceededBoolean := false
	var CookieErrorString1 string
	SourceSearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearchSheetName + "!A4:G5000")
	sourceIterator := 0
	loc, _ := time.LoadLocation("Europe/Rome")
	currentTime := time.Now().In(loc)
	Time := currentTime.Format("2006-01-02")
	for sourceIterator < len(SourceSearchQueryFromNOS) {
		if len(SourceSearchQueryFromNOS[sourceIterator]) < 2 {
			return
		}
		fmt.Println(MinFollower)
		fmt.Println(MaxFollower)
		fmt.Println(MinN)
		fmt.Println(MinNStar)
		fmt.Println(NDelta)
		fmt.Println(SessionID)
		userName := SourceSearchQueryFromNOS[sourceIterator][0]
		if userName == "" {
			sugar.Infof("queryString for search is nil ")
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetStatusCode(200)
			ctx.SetBody([]byte("Failed! Unable to Find USERNAME shared in URL"))
			sugar.Infof("calling ig reprts failure due to username!")
			sourceIterator++
			continue
		}
		LastFetchedFollowerCount := SourceSearchQueryFromNOS[sourceIterator][1]
		if LastFetchedFollowerCount == "" {
			LastFetchedFollowerCount = "10"
			sourceIterator++
			continue
		}
		fmt.Println(userName)
		fmt.Println(LastFetchedFollowerCount)
		FollowersList, _, LatestFollowerCount := ig.GetNewFollowers(userName, LastFetchedFollowerCount, SessionID.(string))
		var nosLatestFollowerCountRows []interface{}
		if LatestFollowerCount != 0 {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName, LatestFollowerCount)
		} else {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName)
		}
		nosLatestFollowerCountFinalValues = append(nosLatestFollowerCountFinalValues, nosLatestFollowerCountRows)
		fmt.Println(FollowersList)
		if len(FollowersList) < 1 {
			sourceIterator++
			continue
		}
		SearchQuery := make(map[string]int)
		if MinFollower != "" {
			temp := MinFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinFollower"] = tempInt
			}
		}
		if MaxFollower != "" {
			temp := MaxFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MaxFollower"] = tempInt
			}
		}
		if MinN != "" {
			temp := MinN
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinN"] = tempInt
			}
		}
		if MinNStar != "" {
			temp := MinNStar
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinNStar"] = tempInt
			}
		}
		NDeltaFloat := 0.0
		if NDelta != "" {
			temp := NDelta
			tempFloat, e := strconv.ParseFloat(temp, 2)
			if e == nil {
				NDeltaFloat = tempFloat
			}
		}
		var reportValues [][]interface{}
		reportValues, _, CookieErrorString1 = ig.GetIGReportNew(FollowersList, SearchQuery, SessionID.(string), NDeltaFloat)
		fmt.Println(reportValues)
		i := 0
		for i < len(reportValues) {
			var searchRow []interface{}
			var dashboardRow []interface{}
			if (len(reportValues[i])) > 5 {
				dashboardRow = append(dashboardRow, Time, "#1", SourceSearchQueryFromNOS[sourceIterator][2], SourceSearchQueryFromNOS[sourceIterator][3], SourceSearchQueryFromNOS[sourceIterator][4], SourceSearchQueryFromNOS[sourceIterator][5], SourceSearchQueryFromNOS[sourceIterator][6], reportValues[i][0], reportValues[i][1], reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6], reportValues[i][7], reportValues[i][8])
				nosDashboardFinalValues = append(nosDashboardFinalValues, dashboardRow)
				searchRow = append(searchRow, Time, reportValues[i][0], userName, reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6])
				nosSearchFinalValues = append(nosSearchFinalValues, searchRow)
			}
			i++
		}
		sourceIterator++
	}
	fmt.Println("*********")
	fmt.Println(nosDashboardFinalValues)
	fmt.Println("#########")
	fmt.Println(nosSearchFinalValues)
	googleSheets.BatchAppend(configs.Configurations.NOSDashboardSheetName, nosDashboardFinalValues)
	existingRows := googleSheets.BatchGet(configs.Configurations.NOSSearchSheetName + "!H4:N5000")
	StartingRow := len(existingRows) + 3 + 1
	googleSheets.BatchWrite(configs.Configurations.NOSSearchSheetName+"!H"+strconv.Itoa(StartingRow)+":N5000", nosSearchFinalValues)
	googleSheets.BatchWrite(configs.Configurations.NOSSearchSheetName+"!A4:B5000", nosLatestFollowerCountFinalValues)
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetStatusCode(200)
	ctx.SetBody([]byte("Success Google Sheet Updated" + " -- " + CookieErrorString1))
	sugar.Infof("calling ig research reports success!")
	// if NoOneSucceededBoolean {
	// 	ctx.Response.Header.Set("Content-Type", "application/json")
	// 	ctx.Response.SetStatusCode(200)
	// 	ctx.SetBody([]byte("Noone passed the filter search query"))
	// 	sugar.Infof("calling ig research reports success!" + " -- " + CookieErrorString1)
	// } else {
	// 	ctx.Response.Header.Set("Content-Type", "application/json")
	// 	ctx.Response.SetStatusCode(200)
	// 	ctx.SetBody([]byte("Something went wrong, not able to fetch data"))
	// 	sugar.Infof("calling ig research reports failure!" + " -- " + CookieErrorString1)
	// }
}

func handleSaveIGResearchToSheets(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a Save IG research request to Google Sheets!")
	userName := ctx.UserValue("USERNAME")
	if userName == nil {
		sugar.Infof("queryString for search is nil ")
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Failed! Unable to Find USERNAME shared in URL"))
		sugar.Infof("calling ig reprts failure due to username!")
		return
	}
	LatestFollowerCount := ctx.UserValue("LatestFollowerCount")
	if LatestFollowerCount == nil {
		LatestFollowerCount = ""
	}
	MinFollower := ctx.UserValue("MinFollower")
	MaxFollower := ctx.UserValue("MaxFollower")
	MinN := ctx.UserValue("MinN")
	MinNStar := ctx.UserValue("MinNStar")
	SessionID := ctx.UserValue("SessionID")
	NDelta := ctx.UserValue("NDelta")
	fmt.Println(SessionID)
	if SessionID != nil {
		temp := SessionID.(string)
		temp = temp[1 : len(temp)-1]
		SessionID = temp
	}
	fmt.Println(userName)
	fmt.Println(LatestFollowerCount)
	fmt.Println(MinFollower)
	fmt.Println(MaxFollower)
	fmt.Println(MinN)
	fmt.Println(MinNStar)
	fmt.Println(NDelta)
	FollowersList, CookieErrorString1 := ig.GetFollowers(userName.(string), LatestFollowerCount.(string)[1:len(LatestFollowerCount.(string))-1], SessionID.(string))
	SearchQuery := make(map[string]int)
	if MinFollower != nil {
		temp := MinFollower.(string)
		temp = temp[1 : len(temp)-1]
		tempInt, e := strconv.Atoi(temp)
		if e == nil {
			SearchQuery["MinFollower"] = tempInt
		}
	}
	if MaxFollower != nil {
		temp := MaxFollower.(string)
		temp = temp[1 : len(temp)-1]
		tempInt, e := strconv.Atoi(temp)
		if e == nil {
			SearchQuery["MaxFollower"] = tempInt
		}
	}
	if MinN != nil {
		temp := MinN.(string)
		temp = temp[1 : len(temp)-1]
		tempInt, e := strconv.Atoi(temp)
		if e == nil {
			SearchQuery["MinN"] = tempInt
		}
	}
	if MinNStar != nil {
		temp := MinNStar.(string)
		temp = temp[1 : len(temp)-1]
		tempInt, e := strconv.Atoi(temp)
		if e == nil {
			SearchQuery["MinNStar"] = tempInt
		}
	}
	NDeltaFloat := 0.0
	if NDelta != nil {
		temp := NDelta.(string)
		temp = temp[1 : len(temp)-1]
		tempFloat, e := strconv.ParseFloat(temp, 2)
		if e == nil {
			NDeltaFloat = tempFloat
		}
	}
	finalValues, NoOneSucceededBoolean, CookieErrorString2 := ig.GetIGReportNew(FollowersList, SearchQuery, SessionID.(string), NDeltaFloat)
	fmt.Println("*********")
	fmt.Println(finalValues)
	if len(finalValues) > 0 {
		fmt.Println(finalValues)
		googleSheets.ClearSheet(configs.Configurations.ResearchJRSheetName)
		googleSheets.BatchWrite(configs.Configurations.ResearchJRSheetName, finalValues)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Google Sheet Updated" + " -- " + CookieErrorString1 + " " + CookieErrorString2))
		sugar.Infof("calling ig research reports success!")
	} else if NoOneSucceededBoolean {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Noone passed the filter search query"))
		sugar.Infof("calling ig research reports success!" + " -- " + CookieErrorString1 + " " + CookieErrorString2)
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, not able to fetch data"))
		sugar.Infof("calling ig research reports failure!" + " -- " + CookieErrorString1 + " " + CookieErrorString2)
	}
}

func handleSaveIGReportToSheets(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a Save IG report request to Google Sheets!")
	userName := ctx.UserValue("USERNAME")
	if userName == nil {
		sugar.Infof("queryString for search is nil ")
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Failed! Unable to Find USERNAME shared in URL"))
		sugar.Infof("calling ig reprts failure due to username!")
		return
	}
	Url := "http://www.instagram.com/" + userName.(string) + "/"

	req, err := http.NewRequest("GET", Url, nil)
	if err != nil {
		// handle err
	}
	req.Header.Set("Authority", "www.instagram.com")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.83 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Set("Cookie", "ig_did=2E8DBEA9-6BAB-4214-BE14-3E92C1956C79; mid=X2Cs0AAEAAH4q10wWRKpkOR7Vcxk; csrftoken=85768r6cbvT6MHcJ7JXRjAz30M7ZyWWP; ds_user_id=41670979469; sessionid=41670979469%3AXIijRyjzHto0c7%3A26; rur=PRN;")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	sugar.Infof(Url)
	var filteredString string
	if err != nil {
		sugar.Infof("Api not responding")
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Failed! Something went wrong to fetch details for this User"))
		sugar.Infof("calling ig reprts failure due to api no response!")
		return
	} else {
		actual := strings.Index(string(body), "<script type=\"text/javascript\">window._sharedData")
		if actual != -1 {
			end := strings.Index(string(body), "<script type=\"text/javascript\">window.__initialDataLoaded(window._sharedData);</script>")
			if end != -1 {
				filteredString = (string(body)[actual+len("<script type=\"text/javascript\">window._sharedData")+2 : end-11])
				fmt.Println(filteredString)
			} else {
				sugar.Infof("-1 While finding json on profile")
				ctx.Response.Header.Set("Content-Type", "application/json")
				ctx.Response.SetStatusCode(200)
				ctx.SetBody([]byte("Failed! Something went wrong to fetch details for this User"))
				sugar.Infof("calling ig reprts failure due to api no response!")
				return
			}
		} else {
			sugar.Infof("-1 While finding json on profile")
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetStatusCode(200)
			ctx.SetBody([]byte("Failed! Something went wrong to fetch details for this User"))
			sugar.Infof("calling ig reprts failure due to api no response!")
			return
		}
		// if actual < 1000 && end < 1000 {
		// 	sugar.Infof("queryString for search is nil ")
		// 	ctx.Response.Header.Set("Content-Type", "application/json")
		// 	ctx.Response.SetStatusCode(200)
		// 	ctx.SetBody([]byte("Failed! Unable to Find USERNAME shared in URL"))
		// 	sugar.Infof("calling ig reprts failure due to username!")
		// 	return
		// }
		fo, err := os.Create("uploads/output.json")
		if err != nil {
			sugar.Infof("Unable to create file")
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetStatusCode(200)
			ctx.SetBody([]byte("Failed! Something went wrong to fetch details for this User"))
			sugar.Infof("calling ig reprts failure due to api no response!")
			return
		}
		defer func() {
			if err := fo.Close(); err != nil {
				sugar.Infof("Unable to close file")
				ctx.Response.Header.Set("Content-Type", "application/json")
				ctx.Response.SetStatusCode(200)
				ctx.SetBody([]byte("Failed! Something went wrong to fetch details for this User"))
				sugar.Infof("calling ig reprts failure due to api no response!")
				return
			}
		}()
		w := bufio.NewWriter(fo)
		if _, err := w.Write([]byte(filteredString)); err != nil {
			sugar.Infof("Unable to read file")
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetStatusCode(200)
			ctx.SetBody([]byte("Failed! Something went wrong to fetch details for this User"))
			sugar.Infof("calling ig reprts failure due to api no response!")
			return
		}
		if err = w.Flush(); err != nil {
			sugar.Infof("Unable to flush file")
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetStatusCode(200)
			ctx.SetBody([]byte("Failed! Something went wrong to fetch details for this User"))
			sugar.Infof("calling ig reprts failure due to api no response!")
			return
		}
	}
	finalValues := ig.GetReport(userName.(string))
	if len(finalValues) > 0 {
		googleSheets.BatchAppend(configs.Configurations.SheetNameWithRange, finalValues)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Google Sheet Updated"))
		sugar.Infof("calling ig reprts success!")
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, not able to fetch data"))
		sugar.Infof("calling ig reprts failure!")
	}
}

func handleSaveIGReportToSheetsNew(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a Save IG report request to Google Sheets!")
	userName := ctx.UserValue("USERNAME")
	if userName == nil {
		sugar.Infof("queryString for search is nil ")
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Failed! Unable to Find USERNAME shared in URL"))
		sugar.Infof("calling ig reprts failure due to username!")
		return
	}
	SessionID := ctx.UserValue("SessionID")
	if SessionID != nil {
		temp := SessionID.(string)
		temp = temp[1 : len(temp)-1]
		SessionID = temp
	}

	finalValues, CookieErrorString := ig.GetReportNew(userName.(string), SessionID.(string))
	LatestIGRAtRow := 3
	currentValueInSheets := googleSheets.BatchGet(configs.Configurations.SheetNameWithRange + "!A1:M50000")
	if len(currentValueInSheets) > 0 {
		LatestIGRAtRow = len(currentValueInSheets) + 1
	}
	fmt.Println(LatestIGRAtRow)
	if len(finalValues) > 0 {
		if LatestIGRAtRow == 3 {
			googleSheets.BatchAppend(configs.Configurations.SheetNameWithRange, finalValues)
		} else {
			googleSheets.BatchWrite(configs.Configurations.SheetNameWithRange+"!A"+strconv.Itoa(LatestIGRAtRow)+":M"+strconv.Itoa(LatestIGRAtRow), finalValues)
		}
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Google Sheet Updated " + CookieErrorString))
		sugar.Infof("calling ig reprts success!")
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, not able to fetch data " + CookieErrorString))
		sugar.Infof("calling ig reprts failure!")
	}
}

func handleIGRDatabaseBackup(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a IGR Database backup request to Google Sheets!")
	currentValueInSheets := googleSheets.BatchGet(configs.Configurations.SheetNameWithRange + "!A4:N50000")
	var finalValuesToSheets [][]interface{}
	if len(currentValueInSheets) > 0 {
		iter := 0
		for iter < len(currentValueInSheets) {
			iter2 := 0
			var row []interface{}
			for iter2 < len(currentValueInSheets[iter]) {
				row = append(row, currentValueInSheets[iter][iter2])
				iter2++
			}
			finalValuesToSheets = append(finalValuesToSheets, row)
			iter++
		}
		googleSheets.BatchAppend(configs.Configurations.IGRDatabaseSheetName, finalValuesToSheets)
	}
	if len(finalValuesToSheets) > 0 {
		googleSheets.ClearSheet("SCAN!A4:M50000")
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success IGR Database Google Sheet Updated "))
		sugar.Infof("calling ig reprts success!")
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, no data to update IGR database"))
		sugar.Infof("calling ig reprts failure!")
	}
}

func handleGetAccountDatabase(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a IGR Database backup request to Google Sheets!")

	userName := ctx.UserValue("USERNAME")
	if userName == nil {
		sugar.Infof("queryString for search is nil ")
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Failed! Unable to Find USERNAME shared in URL"))
		sugar.Infof("calling failure due to no username!")
		return
	}
	SessionID := ctx.UserValue("SessionID")
	if SessionID != nil {
		temp := SessionID.(string)
		temp = temp[1 : len(temp)-1]
		SessionID = temp
	}
	finalValuesToSheets, _ := ig.GetAccountFollowersDetails(userName.(string), "500", SessionID.(string))
	if len(finalValuesToSheets) > 0 {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Account Database Google Sheet Updated "))
		sugar.Infof("calling ig reprts success!")
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, please check account is accessible from the mentioned SessionID"))
		sugar.Infof("calling ig reprts failure!")
	}
}

func handleNOSSearchSetup7(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a NOS Search request to Google Sheets!")
	SearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch7SheetName + "!A2:N2")
	fmt.Println(SearchQueryFromNOS)
	var nosSearchFinalValues [][]interface{}
	var nosDashboardFinalValues [][]interface{}
	var nosLatestFollowerCountFinalValues [][]interface{}
	var MinFollower string
	var MaxFollower string
	var MinN string
	var MinNStar string
	var NDelta string

	if len(SearchQueryFromNOS) == 1 {
		if len(SearchQueryFromNOS[0]) > 9 {
			MinFollower = SearchQueryFromNOS[0][8]
			MaxFollower = SearchQueryFromNOS[0][10]
			MinFollower = strings.Replace(MinFollower, ",", "", -1)
			MaxFollower = strings.Replace(MaxFollower, ",", "", -1)
			if len(SearchQueryFromNOS[0]) > 11 {
				MinN = SearchQueryFromNOS[0][11]
				MinN = strings.Replace(MinN, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 12 {
				MinNStar = SearchQueryFromNOS[0][12]
				MinNStar = strings.Replace(MinNStar, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 13 {
				NDelta = SearchQueryFromNOS[0][13]
				NDelta = strings.Replace(NDelta, ",", "", -1)
			}
		}
	}

	SessionID := ctx.UserValue("SessionID")
	if SessionID != nil {
		temp := SessionID.(string)
		temp = temp[1 : len(temp)-1]
		SessionID = temp
		CookieFinder := googleSheets.BatchGet(configs.Configurations.CookieFinderSheet + "!A9:B9")
		if len(CookieFinder) == 1 {
			if len(CookieFinder[0]) == 2 {
				SessionID = CookieFinder[0][1]
				fmt.Print("Received Session ids from Sheet: ")
				fmt.Println(SessionID.(string))
			}
		}
	}
	NoOneSucceededBoolean := false
	var CookieErrorString1 string
	SourceSearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch7SheetName + "!A4:G5000")
	sourceIterator := 0
	loc, _ := time.LoadLocation("Europe/Rome")
	currentTime := time.Now().In(loc)
	Time := currentTime.Format("2006-01-02")
	for sourceIterator < len(SourceSearchQueryFromNOS) {
		if len(SourceSearchQueryFromNOS[sourceIterator]) < 2 {
			sourceIterator++
			continue
		}
		fmt.Println(MinFollower)
		fmt.Println(MaxFollower)
		fmt.Println(MinN)
		fmt.Println(MinNStar)
		fmt.Println(NDelta)
		fmt.Println(SessionID)
		userName := SourceSearchQueryFromNOS[sourceIterator][0]
		if userName == "" {
			sugar.Infof("queryString for search is nil ")
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetStatusCode(200)
			ctx.SetBody([]byte("Failed! Unable to Find USERNAME shared in URL"))
			sugar.Infof("calling ig reprts failure due to username!")
			sourceIterator++
			continue
		}
		LastFetchedFollowerCount := SourceSearchQueryFromNOS[sourceIterator][1]
		if LastFetchedFollowerCount == "" {
			LastFetchedFollowerCount = "10"
		}
		fmt.Println(userName)
		fmt.Println(LastFetchedFollowerCount)
		FollowersList, _, LatestFollowerCount := ig.GetNewFollowers(userName, LastFetchedFollowerCount, SessionID.(string))
		var nosLatestFollowerCountRows []interface{}
		if LatestFollowerCount != 0 {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName, LatestFollowerCount)
		} else {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName)
		}
		nosLatestFollowerCountFinalValues = append(nosLatestFollowerCountFinalValues, nosLatestFollowerCountRows)
		fmt.Println(FollowersList)
		SearchQuery := make(map[string]int)
		if MinFollower != "" {
			temp := MinFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinFollower"] = tempInt
			}
		}
		if MaxFollower != "" {
			temp := MaxFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MaxFollower"] = tempInt
			}
		}
		if MinN != "" {
			temp := MinN
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinN"] = tempInt
			}
		}
		if MinNStar != "" {
			temp := MinNStar
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinNStar"] = tempInt
			}
		}
		NDeltaFloat := 0.0
		if NDelta != "" {
			temp := NDelta
			tempFloat, e := strconv.ParseFloat(temp, 2)
			if e == nil {
				NDeltaFloat = tempFloat
			}
		}
		var reportValues [][]interface{}
		reportValues, NoOneSucceededBoolean, CookieErrorString1 = ig.GetIGReportNew(FollowersList, SearchQuery, SessionID.(string), NDeltaFloat)
		fmt.Println(reportValues)
		i := 0
		for i < len(reportValues) {
			var searchRow []interface{}
			var dashboardRow []interface{}
			if (len(reportValues[i])) > 5 {
				dashboardRow = append(dashboardRow, Time, "#7", SourceSearchQueryFromNOS[sourceIterator][2], SourceSearchQueryFromNOS[sourceIterator][3], SourceSearchQueryFromNOS[sourceIterator][4], SourceSearchQueryFromNOS[sourceIterator][5], SourceSearchQueryFromNOS[sourceIterator][6], reportValues[i][0], reportValues[i][1], reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6], reportValues[i][7], reportValues[i][8])
				nosDashboardFinalValues = append(nosDashboardFinalValues, dashboardRow)
				searchRow = append(searchRow, Time, reportValues[i][0], userName, reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6])
				nosSearchFinalValues = append(nosSearchFinalValues, searchRow)
			}
			i++
		}
		sourceIterator++
	}
	fmt.Println("*********")
	fmt.Println(nosDashboardFinalValues)
	fmt.Println("#########")
	fmt.Println(nosSearchFinalValues)
	if len(nosDashboardFinalValues) > 0 {
		googleSheets.BatchAppend(configs.Configurations.NOSDashboardSheetName, nosDashboardFinalValues)
		existingRows := googleSheets.BatchGet(configs.Configurations.NOSSearch7SheetName + "!H4:N5000")
		StartingRow := len(existingRows) + 3 + 1
		googleSheets.BatchWrite(configs.Configurations.NOSSearch7SheetName+"!H"+strconv.Itoa(StartingRow)+":N5000", nosSearchFinalValues)
		googleSheets.BatchWrite(configs.Configurations.NOSSearch7SheetName+"!A4:B5000", nosLatestFollowerCountFinalValues)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Google Sheet Updated" + " -- " + CookieErrorString1))
		sugar.Infof("calling ig research reports success!")
	} else if NoOneSucceededBoolean {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Noone passed the filter search query"))
		sugar.Infof("calling ig research reports success!" + " -- " + CookieErrorString1)
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, not able to fetch data"))
		sugar.Infof("calling ig research reports failure!" + " -- " + CookieErrorString1)
	}
}

func handleNOSSearchSetup8(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a NOS Search request to Google Sheets!")
	SearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch8SheetName + "!A2:N2")
	fmt.Println(SearchQueryFromNOS)
	var nosSearchFinalValues [][]interface{}
	var nosDashboardFinalValues [][]interface{}
	var nosLatestFollowerCountFinalValues [][]interface{}
	var MinFollower string
	var MaxFollower string
	var MinN string
	var MinNStar string
	var NDelta string

	if len(SearchQueryFromNOS) == 1 {
		if len(SearchQueryFromNOS[0]) > 9 {
			MinFollower = SearchQueryFromNOS[0][8]
			MaxFollower = SearchQueryFromNOS[0][10]
			MinFollower = strings.Replace(MinFollower, ",", "", -1)
			MaxFollower = strings.Replace(MaxFollower, ",", "", -1)
			if len(SearchQueryFromNOS[0]) > 11 {
				MinN = SearchQueryFromNOS[0][11]
				MinN = strings.Replace(MinN, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 12 {
				MinNStar = SearchQueryFromNOS[0][12]
				MinNStar = strings.Replace(MinNStar, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 13 {
				NDelta = SearchQueryFromNOS[0][13]
				NDelta = strings.Replace(NDelta, ",", "", -1)
			}
		}
	}

	SessionID := ctx.UserValue("SessionID")
	if SessionID != nil {
		temp := SessionID.(string)
		temp = temp[1 : len(temp)-1]
		SessionID = temp
		CookieFinder := googleSheets.BatchGet(configs.Configurations.CookieFinderSheet + "!A10:B10")
		if len(CookieFinder) == 1 {
			if len(CookieFinder[0]) == 2 {
				SessionID = CookieFinder[0][1]
				fmt.Print("Received Session ids from Sheet: ")
				fmt.Println(SessionID.(string))
			}
		}
	}
	NoOneSucceededBoolean := false
	var CookieErrorString1 string
	SourceSearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch8SheetName + "!A4:G5000")
	sourceIterator := 0
	loc, _ := time.LoadLocation("Europe/Rome")
	currentTime := time.Now().In(loc)
	Time := currentTime.Format("2006-01-02")
	for sourceIterator < len(SourceSearchQueryFromNOS) {
		if len(SourceSearchQueryFromNOS[sourceIterator]) < 2 {
			sourceIterator++
			continue
		}
		fmt.Println(MinFollower)
		fmt.Println(MaxFollower)
		fmt.Println(MinN)
		fmt.Println(MinNStar)
		fmt.Println(NDelta)
		fmt.Println(SessionID)
		userName := SourceSearchQueryFromNOS[sourceIterator][0]
		if userName == "" {
			sugar.Infof("queryString for search is nil ")
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetStatusCode(200)
			ctx.SetBody([]byte("Failed! Unable to Find USERNAME shared in URL"))
			sugar.Infof("calling ig reprts failure due to username!")
			sourceIterator++
			continue
		}
		LastFetchedFollowerCount := SourceSearchQueryFromNOS[sourceIterator][1]
		if LastFetchedFollowerCount == "" {
			LastFetchedFollowerCount = "10"
		}
		fmt.Println(userName)
		fmt.Println(LastFetchedFollowerCount)
		FollowersList, _, LatestFollowerCount := ig.GetNewFollowers(userName, LastFetchedFollowerCount, SessionID.(string))
		var nosLatestFollowerCountRows []interface{}
		if LatestFollowerCount != 0 {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName, LatestFollowerCount)
		} else {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName)
		}
		nosLatestFollowerCountFinalValues = append(nosLatestFollowerCountFinalValues, nosLatestFollowerCountRows)
		fmt.Println(FollowersList)
		SearchQuery := make(map[string]int)
		if MinFollower != "" {
			temp := MinFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinFollower"] = tempInt
			}
		}
		if MaxFollower != "" {
			temp := MaxFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MaxFollower"] = tempInt
			}
		}
		if MinN != "" {
			temp := MinN
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinN"] = tempInt
			}
		}
		if MinNStar != "" {
			temp := MinNStar
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinNStar"] = tempInt
			}
		}
		NDeltaFloat := 0.0
		if NDelta != "" {
			temp := NDelta
			tempFloat, e := strconv.ParseFloat(temp, 2)
			if e == nil {
				NDeltaFloat = tempFloat
			}
		}
		var reportValues [][]interface{}
		reportValues, NoOneSucceededBoolean, CookieErrorString1 = ig.GetIGReportNew(FollowersList, SearchQuery, SessionID.(string), NDeltaFloat)
		fmt.Println(reportValues)
		i := 0
		for i < len(reportValues) {
			var searchRow []interface{}
			var dashboardRow []interface{}
			if (len(reportValues[i])) > 5 {
				dashboardRow = append(dashboardRow, Time, "#8", SourceSearchQueryFromNOS[sourceIterator][2], SourceSearchQueryFromNOS[sourceIterator][3], SourceSearchQueryFromNOS[sourceIterator][4], SourceSearchQueryFromNOS[sourceIterator][5], SourceSearchQueryFromNOS[sourceIterator][6], reportValues[i][0], reportValues[i][1], reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6], reportValues[i][7], reportValues[i][8])
				nosDashboardFinalValues = append(nosDashboardFinalValues, dashboardRow)
				searchRow = append(searchRow, Time, reportValues[i][0], userName, reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6])
				nosSearchFinalValues = append(nosSearchFinalValues, searchRow)
			}
			i++
		}
		sourceIterator++
	}
	fmt.Println("*********")
	fmt.Println(nosDashboardFinalValues)
	fmt.Println("#########")
	fmt.Println(nosSearchFinalValues)
	if len(nosDashboardFinalValues) > 0 {
		googleSheets.BatchAppend(configs.Configurations.NOSDashboardSheetName, nosDashboardFinalValues)
		existingRows := googleSheets.BatchGet(configs.Configurations.NOSSearch8SheetName + "!H4:N5000")
		StartingRow := len(existingRows) + 3 + 1
		googleSheets.BatchWrite(configs.Configurations.NOSSearch8SheetName+"!H"+strconv.Itoa(StartingRow)+":N5000", nosSearchFinalValues)
		googleSheets.BatchWrite(configs.Configurations.NOSSearch8SheetName+"!A4:B5000", nosLatestFollowerCountFinalValues)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Google Sheet Updated" + " -- " + CookieErrorString1))
		sugar.Infof("calling ig research reports success!")
	} else if NoOneSucceededBoolean {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Noone passed the filter search query"))
		sugar.Infof("calling ig research reports success!" + " -- " + CookieErrorString1)
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, not able to fetch data"))
		sugar.Infof("calling ig research reports failure!" + " -- " + CookieErrorString1)
	}
}

func handleNOSSearchSetup9(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a NOS Search request to Google Sheets!")
	SearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch9SheetName + "!A2:N2")
	fmt.Println(SearchQueryFromNOS)
	var nosSearchFinalValues [][]interface{}
	var nosDashboardFinalValues [][]interface{}
	var nosLatestFollowerCountFinalValues [][]interface{}
	var MinFollower string
	var MaxFollower string
	var MinN string
	var MinNStar string
	var NDelta string

	if len(SearchQueryFromNOS) == 1 {
		if len(SearchQueryFromNOS[0]) > 9 {
			MinFollower = SearchQueryFromNOS[0][8]
			MaxFollower = SearchQueryFromNOS[0][10]
			MinFollower = strings.Replace(MinFollower, ",", "", -1)
			MaxFollower = strings.Replace(MaxFollower, ",", "", -1)
			if len(SearchQueryFromNOS[0]) > 11 {
				MinN = SearchQueryFromNOS[0][11]
				MinN = strings.Replace(MinN, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 12 {
				MinNStar = SearchQueryFromNOS[0][12]
				MinNStar = strings.Replace(MinNStar, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 13 {
				NDelta = SearchQueryFromNOS[0][13]
				NDelta = strings.Replace(NDelta, ",", "", -1)
			}
		}
	}

	SessionID := ctx.UserValue("SessionID")
	if SessionID != nil {
		temp := SessionID.(string)
		temp = temp[1 : len(temp)-1]
		SessionID = temp
		CookieFinder := googleSheets.BatchGet(configs.Configurations.CookieFinderSheet + "!A11:B11")
		if len(CookieFinder) == 1 {
			if len(CookieFinder[0]) == 2 {
				SessionID = CookieFinder[0][1]
				fmt.Print("Received Session ids from Sheet: ")
				fmt.Println(SessionID.(string))
			}
		}
	}
	NoOneSucceededBoolean := false
	var CookieErrorString1 string
	SourceSearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch9SheetName + "!A4:G5000")
	sourceIterator := 0
	loc, _ := time.LoadLocation("Europe/Rome")
	currentTime := time.Now().In(loc)
	Time := currentTime.Format("2006-01-02")
	for sourceIterator < len(SourceSearchQueryFromNOS) {
		if len(SourceSearchQueryFromNOS[sourceIterator]) < 2 {
			sourceIterator++
			continue
		}
		fmt.Println(MinFollower)
		fmt.Println(MaxFollower)
		fmt.Println(MinN)
		fmt.Println(MinNStar)
		fmt.Println(NDelta)
		fmt.Println(SessionID)
		userName := SourceSearchQueryFromNOS[sourceIterator][0]
		if userName == "" {
			sugar.Infof("queryString for search is nil ")
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetStatusCode(200)
			ctx.SetBody([]byte("Failed! Unable to Find USERNAME shared in URL"))
			sugar.Infof("calling ig reprts failure due to username!")
			sourceIterator++
			continue
		}
		LastFetchedFollowerCount := SourceSearchQueryFromNOS[sourceIterator][1]
		if LastFetchedFollowerCount == "" {
			LastFetchedFollowerCount = "10"
		}
		fmt.Println(userName)
		fmt.Println(LastFetchedFollowerCount)
		FollowersList, _, LatestFollowerCount := ig.GetNewFollowers(userName, LastFetchedFollowerCount, SessionID.(string))
		var nosLatestFollowerCountRows []interface{}
		if LatestFollowerCount != 0 {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName, LatestFollowerCount)
		} else {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName)
		}
		nosLatestFollowerCountFinalValues = append(nosLatestFollowerCountFinalValues, nosLatestFollowerCountRows)
		fmt.Println(FollowersList)
		SearchQuery := make(map[string]int)
		if MinFollower != "" {
			temp := MinFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinFollower"] = tempInt
			}
		}
		if MaxFollower != "" {
			temp := MaxFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MaxFollower"] = tempInt
			}
		}
		if MinN != "" {
			temp := MinN
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinN"] = tempInt
			}
		}
		if MinNStar != "" {
			temp := MinNStar
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinNStar"] = tempInt
			}
		}
		NDeltaFloat := 0.0
		if NDelta != "" {
			temp := NDelta
			tempFloat, e := strconv.ParseFloat(temp, 2)
			if e == nil {
				NDeltaFloat = tempFloat
			}
		}
		var reportValues [][]interface{}
		reportValues, NoOneSucceededBoolean, CookieErrorString1 = ig.GetIGReportNew(FollowersList, SearchQuery, SessionID.(string), NDeltaFloat)
		fmt.Println(reportValues)
		i := 0
		for i < len(reportValues) {
			var searchRow []interface{}
			var dashboardRow []interface{}
			if (len(reportValues[i])) > 5 {
				dashboardRow = append(dashboardRow, Time, "#9", SourceSearchQueryFromNOS[sourceIterator][2], SourceSearchQueryFromNOS[sourceIterator][3], SourceSearchQueryFromNOS[sourceIterator][4], SourceSearchQueryFromNOS[sourceIterator][5], SourceSearchQueryFromNOS[sourceIterator][6], reportValues[i][0], reportValues[i][1], reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6], reportValues[i][7], reportValues[i][8])
				nosDashboardFinalValues = append(nosDashboardFinalValues, dashboardRow)
				searchRow = append(searchRow, Time, reportValues[i][0], userName, reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6])
				nosSearchFinalValues = append(nosSearchFinalValues, searchRow)
			}
			i++
		}
		sourceIterator++
	}
	fmt.Println("*********")
	fmt.Println(nosDashboardFinalValues)
	fmt.Println("#########")
	fmt.Println(nosSearchFinalValues)
	if len(nosDashboardFinalValues) > 0 {
		googleSheets.BatchAppend(configs.Configurations.NOSDashboardSheetName, nosDashboardFinalValues)
		existingRows := googleSheets.BatchGet(configs.Configurations.NOSSearch9SheetName + "!H4:N5000")
		StartingRow := len(existingRows) + 3 + 1
		googleSheets.BatchWrite(configs.Configurations.NOSSearch9SheetName+"!H"+strconv.Itoa(StartingRow)+":N5000", nosSearchFinalValues)
		googleSheets.BatchWrite(configs.Configurations.NOSSearch9SheetName+"!A4:B5000", nosLatestFollowerCountFinalValues)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Google Sheet Updated" + " -- " + CookieErrorString1))
		sugar.Infof("calling ig research reports success!")
	} else if NoOneSucceededBoolean {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Noone passed the filter search query"))
		sugar.Infof("calling ig research reports success!" + " -- " + CookieErrorString1)
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, not able to fetch data"))
		sugar.Infof("calling ig research reports failure!" + " -- " + CookieErrorString1)
	}
}

func handleNOSSearchSetup10(ctx *fasthttp.RequestCtx) {
	configs.SetConfig()
	sugar.Infof("received a NOS Search request to Google Sheets!")
	SearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch10SheetName + "!A2:N2")
	fmt.Println(SearchQueryFromNOS)
	var nosSearchFinalValues [][]interface{}
	var nosDashboardFinalValues [][]interface{}
	var nosLatestFollowerCountFinalValues [][]interface{}
	var MinFollower string
	var MaxFollower string
	var MinN string
	var MinNStar string
	var NDelta string

	if len(SearchQueryFromNOS) == 1 {
		if len(SearchQueryFromNOS[0]) > 9 {
			MinFollower = SearchQueryFromNOS[0][8]
			MaxFollower = SearchQueryFromNOS[0][10]
			MinFollower = strings.Replace(MinFollower, ",", "", -1)
			MaxFollower = strings.Replace(MaxFollower, ",", "", -1)
			if len(SearchQueryFromNOS[0]) > 11 {
				MinN = SearchQueryFromNOS[0][11]
				MinN = strings.Replace(MinN, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 12 {
				MinNStar = SearchQueryFromNOS[0][12]
				MinNStar = strings.Replace(MinNStar, ",", "", -1)
			}
			if len(SearchQueryFromNOS[0]) > 13 {
				NDelta = SearchQueryFromNOS[0][13]
				NDelta = strings.Replace(NDelta, ",", "", -1)
			}
		}
	}

	SessionID := ctx.UserValue("SessionID")
	if SessionID != nil {
		temp := SessionID.(string)
		temp = temp[1 : len(temp)-1]
		SessionID = temp
		CookieFinder := googleSheets.BatchGet(configs.Configurations.CookieFinderSheet + "!A12:B12")
		if len(CookieFinder) == 1 {
			if len(CookieFinder[0]) == 2 {
				SessionID = CookieFinder[0][1]
				fmt.Print("Received Session ids from Sheet: ")
				fmt.Println(SessionID.(string))
			}
		}
	}
	NoOneSucceededBoolean := false
	var CookieErrorString1 string
	SourceSearchQueryFromNOS := googleSheets.BatchGet(configs.Configurations.NOSSearch10SheetName + "!A4:G5000")
	sourceIterator := 0
	loc, _ := time.LoadLocation("Europe/Rome")
	currentTime := time.Now().In(loc)
	Time := currentTime.Format("2006-01-02")
	for sourceIterator < len(SourceSearchQueryFromNOS) {
		if len(SourceSearchQueryFromNOS[sourceIterator]) < 2 {
			sourceIterator++
			continue
		}
		fmt.Println(MinFollower)
		fmt.Println(MaxFollower)
		fmt.Println(MinN)
		fmt.Println(MinNStar)
		fmt.Println(NDelta)
		fmt.Println(SessionID)
		userName := SourceSearchQueryFromNOS[sourceIterator][0]
		if userName == "" {
			sugar.Infof("queryString for search is nil ")
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetStatusCode(200)
			ctx.SetBody([]byte("Failed! Unable to Find USERNAME shared in URL"))
			sugar.Infof("calling ig reprts failure due to username!")
			sourceIterator++
			continue
		}
		LastFetchedFollowerCount := SourceSearchQueryFromNOS[sourceIterator][1]
		if LastFetchedFollowerCount == "" {
			LastFetchedFollowerCount = "10"
		}
		fmt.Println(userName)
		fmt.Println(LastFetchedFollowerCount)
		FollowersList, _, LatestFollowerCount := ig.GetNewFollowers(userName, LastFetchedFollowerCount, SessionID.(string))
		var nosLatestFollowerCountRows []interface{}
		if LatestFollowerCount != 0 {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName, LatestFollowerCount)
		} else {
			nosLatestFollowerCountRows = append(nosLatestFollowerCountRows, userName)
		}
		nosLatestFollowerCountFinalValues = append(nosLatestFollowerCountFinalValues, nosLatestFollowerCountRows)
		fmt.Println(FollowersList)
		SearchQuery := make(map[string]int)
		if MinFollower != "" {
			temp := MinFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinFollower"] = tempInt
			}
		}
		if MaxFollower != "" {
			temp := MaxFollower
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MaxFollower"] = tempInt
			}
		}
		if MinN != "" {
			temp := MinN
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinN"] = tempInt
			}
		}
		if MinNStar != "" {
			temp := MinNStar
			tempInt, e := strconv.Atoi(temp)
			if e == nil {
				SearchQuery["MinNStar"] = tempInt
			}
		}
		NDeltaFloat := 0.0
		if NDelta != "" {
			temp := NDelta
			tempFloat, e := strconv.ParseFloat(temp, 2)
			if e == nil {
				NDeltaFloat = tempFloat
			}
		}
		var reportValues [][]interface{}
		reportValues, NoOneSucceededBoolean, CookieErrorString1 = ig.GetIGReportNew(FollowersList, SearchQuery, SessionID.(string), NDeltaFloat)
		fmt.Println(reportValues)
		i := 0
		for i < len(reportValues) {
			var searchRow []interface{}
			var dashboardRow []interface{}
			if (len(reportValues[i])) > 5 {
				dashboardRow = append(dashboardRow, Time, "#10", SourceSearchQueryFromNOS[sourceIterator][2], SourceSearchQueryFromNOS[sourceIterator][3], SourceSearchQueryFromNOS[sourceIterator][4], SourceSearchQueryFromNOS[sourceIterator][5], SourceSearchQueryFromNOS[sourceIterator][6], reportValues[i][0], reportValues[i][1], reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6], reportValues[i][7], reportValues[i][8])
				nosDashboardFinalValues = append(nosDashboardFinalValues, dashboardRow)
				searchRow = append(searchRow, Time, reportValues[i][0], userName, reportValues[i][3], reportValues[i][4], reportValues[i][5], reportValues[i][6])
				nosSearchFinalValues = append(nosSearchFinalValues, searchRow)
			}
			i++
		}
		sourceIterator++
	}
	fmt.Println("*********")
	fmt.Println(nosDashboardFinalValues)
	fmt.Println("#########")
	fmt.Println(nosSearchFinalValues)
	if len(nosDashboardFinalValues) > 0 {
		googleSheets.BatchAppend(configs.Configurations.NOSDashboardSheetName, nosDashboardFinalValues)
		existingRows := googleSheets.BatchGet(configs.Configurations.NOSSearch10SheetName + "!H4:N5000")
		StartingRow := len(existingRows) + 3 + 1
		googleSheets.BatchWrite(configs.Configurations.NOSSearch10SheetName+"!H"+strconv.Itoa(StartingRow)+":N5000", nosSearchFinalValues)
		googleSheets.BatchWrite(configs.Configurations.NOSSearch10SheetName+"!A4:B5000", nosLatestFollowerCountFinalValues)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Google Sheet Updated" + " -- " + CookieErrorString1))
		sugar.Infof("calling ig research reports success!")
	} else if NoOneSucceededBoolean {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Noone passed the filter search query"))
		sugar.Infof("calling ig research reports success!" + " -- " + CookieErrorString1)
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, not able to fetch data"))
		sugar.Infof("calling ig research reports failure!" + " -- " + CookieErrorString1)
	}
}
