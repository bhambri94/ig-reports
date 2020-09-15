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
	router.GET("/v1/get/ig/report/username=:USERNAME", handleSaveIGReportToSheets)
	router.GET("/v1/get/ig/research/username=:USERNAME/LatestFollowerCount=:LatestFollowerCount/MinFollower=:MinFollower/MaxFollower=:MaxFollower/MinN=:MinN/MinNStar=:MinNStar", handleSaveIGResearchToSheets)
	log.Fatal(fasthttp.ListenAndServe(":3003", router.Handler))
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

	FollowersList := ig.GetFollowers(userName.(string), LatestFollowerCount.(string)[1:len(LatestFollowerCount.(string))-1])
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
	finalValues, NoOneSucceededBoolean := ig.GetIGReportNew(FollowersList, SearchQuery)
	fmt.Println("*********")
	fmt.Println(finalValues)
	if len(finalValues) > 0 {
		fmt.Println(finalValues)
		googleSheets.ClearSheet(configs.Configurations.ResearchJRSheetName)
		googleSheets.BatchWrite(configs.Configurations.ResearchJRSheetName, finalValues)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Google Sheet Updated"))
		sugar.Infof("calling ig research reports success!")
	} else if NoOneSucceededBoolean {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Noone passed the filter search query"))
		sugar.Infof("calling ig research reports success!")
	} else {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Something went wrong, not able to fetch data"))
		sugar.Infof("calling ig research reports failure!")
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
