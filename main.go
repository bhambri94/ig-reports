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
	finalValues := ig.GetIGReport(FollowersList, SearchQuery)
	fmt.Println("*********")
	fmt.Println(finalValues)
	if len(finalValues) > 0 {
		googleSheets.ClearSheet(configs.Configurations.ResearchJRSheetName)
		googleSheets.BatchWrite(configs.Configurations.ResearchJRSheetName, finalValues)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Success Google Sheet Updated"))
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
	req.Header.Set("Cookie", "mid=XSMB8QAEAAEs3mQemNZLh2dhx98f; ig_did=EBB71BE2-8122-414C-9E28-4946DF598A00; datr=mgUcXzN0FK6UZc2wKHzFVdS8; fbm_124024574287414=base_domain=.instagram.com; shbid=18143; shbts=1599896664.3834596; rur=ATN; fbsr_124024574287414=Gx2jo4u1YNucR8uhdoS_OZz11ssN3ZH1Hm99OsmpsrE.eyJ1c2VyX2lkIjoiMTAwMDAyMDg2MzA5OTAzIiwiY29kZSI6IkFRREFOUlJ5VUtRN3BwaENSY1BGOVNLM3hvb1hVb2Q2UFBDMkJpaGtuUkVnd3BPQTZuc1dSVHdUbzNQUldxdjFnZi1nM0EwMlhDY01rZE9qQ1lzMzB1Z19KanZVUENVc2d5YzhFcml2cUNGU3pmOERaUUpRSXVFc2NxTGNNeWw0WlU2dURidHdZekRQWFJHQUhnQ0g0bkNvb3R0NnZyUWFkdGJ0SWV0d3BwcnZNc2hTbUJidmNab2tndkVxd3h1N3Jyd3FrU2F0OGdiT0xYWG1rV3p1T2QzT2tLUVRVdlBXc2xWOHpRellwal9sbjMzVjZPb0tFNmZMNm9TVnhNZk0wdU1aanprWDFPZ0IweFlmYU1PcHJGMW9qcUFmakJUMGJGUVl2LWVuZlNYeHIyS29uVS1LTzJrdl96TU9jVVNhTm5KeDJLVDN6NTcyMUhxenIxMlUtZDBRIiwib2F1dGhfdG9rZW4iOiJFQUFCd3pMaXhuallCQU0xbXMxQXJMMXQ2cU5rWGw5RTFnOXYyWkFyUHRyVk9wMHVFZmhHY05HMTJOaFpCZUhRNlFWalI5em1OcU85RU1Bc2pWOE85N2FXcVpCM2tQdkRaQncyb0gyelNrd3Azc0xRMXVMZ2p0OXlLOXE1WkN6QlBXaTRqdTl5bmtvT201NHB1d0s5MTFFSHdvSGgwZlZIWkNaQXkzSXdRa3hrT2NpbkRYZ0RzVzBaQiIsImFsZ29yaXRobSI6IkhNQUMtU0hBMjU2IiwiaXNzdWVkX2F0IjoxNTk5OTEwMzQwfQ; csrftoken=AYPfDg0kdLFbPNZbVHkkJIojQ1wPKSdH; ds_user_id=41309535897; sessionid=41309535897%3A4XfannYCtGdfzr%3A29;")

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
