package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/anaskhan96/soup"
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
	log.Fatal(fasthttp.ListenAndServe(":8011", router.Handler))
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
	sugar.Infof(Url)
	resp, err := soup.Get(Url)
	if err != nil {
		sugar.Infof("Api not responding")
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetStatusCode(200)
		ctx.SetBody([]byte("Failed! Something went wrong to fetch details for this User"))
		sugar.Infof("calling ig reprts failure due to api no response!")
		return
	} else {
		actual := strings.Index(string(resp), "<script type=\"text/javascript\">window._sharedData")
		end := strings.Index(string(resp), "<script type=\"text/javascript\">window.__initialDataLoaded(window._sharedData);</script>")
		if actual < 1000 && end < 1000 {
			sugar.Infof("queryString for search is nil ")
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetStatusCode(200)
			ctx.SetBody([]byte("Failed! Unable to Find USERNAME shared in URL"))
			sugar.Infof("calling ig reprts failure due to username!")
			return
		}
		filteredString := (string(resp)[actual+len("<script type=\"text/javascript\">window._sharedData")+2 : end-11])
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
	}
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetStatusCode(200)
	ctx.SetBody([]byte("Success Google Sheet Updated"))
	sugar.Infof("calling ig reprts success!")
	// sugar.Infof(string(ctx.Request.Body()))
}
