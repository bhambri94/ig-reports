package ig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
)

func GetReport(userName string) [][]interface{} {
	var finalValues [][]interface{}
	NumberOfPosts30Days := 0
	NumberOfPosts90Days := 0
	NumberOfPosts180Days := 0
	jsonFile, err := os.Open("uploads/output.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened output.json")
	byteValue, _ := ioutil.ReadAll(jsonFile)
	fmt.Println(string(byteValue))
	defer jsonFile.Close()
	var igResponse IGResponse
	json.Unmarshal(byteValue, &igResponse)
	loc, _ := time.LoadLocation("Europe/Rome")
	currentTime := time.Now().In(loc)
	Time := currentTime.Format("2006-01-02")
	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	timestamp = timestamp[:10]
	t, _ := strconv.Atoi(timestamp)
	var row []interface{}
	TotalLikes := 0.0
	TotalComments := 0.0
	if len(igResponse.EntryData.ProfilePage) > 0 {
		userId := igResponse.EntryData.ProfilePage[0].Graphql.User.ID
		EndCursor := igResponse.EntryData.ProfilePage[0].Graphql.User.EdgeOwnerToTimelineMedia.PageInfo.EndCursor
		row = append(row, Time)
		row = append(row, igResponse.EntryData.ProfilePage[0].Graphql.User.Username)
		Followers := igResponse.EntryData.ProfilePage[0].Graphql.User.EdgeFollowedBy.Count
		row = append(row, Followers)
		// row = append(row, "NA")
		row = append(row, igResponse.EntryData.ProfilePage[0].Graphql.User.EdgeOwnerToTimelineMedia.Count)

		i := 0
		Engagement := make([]int, 12)
		for i < 12 && i < len(igResponse.EntryData.ProfilePage[0].Graphql.User.EdgeOwnerToTimelineMedia.Edges) {
			Likes := igResponse.EntryData.ProfilePage[0].Graphql.User.EdgeOwnerToTimelineMedia.Edges[i].Node.EdgeLikedBy.Count
			Comments := igResponse.EntryData.ProfilePage[0].Graphql.User.EdgeOwnerToTimelineMedia.Edges[i].Node.EdgeMediaToComment.Count
			TotalLikes = TotalLikes + float64(Likes)
			TotalComments = TotalComments + float64(Comments)
			Engagement[i] = (Likes + Comments)
			MediaTimestamp := igResponse.EntryData.ProfilePage[0].Graphql.User.EdgeOwnerToTimelineMedia.Edges[i].Node.TakenAtTimestamp
			if t-MediaTimestamp < (30 * 24 * 60 * 60) {
				NumberOfPosts30Days++
			}
			i++
		}
		sort.Sort(sort.IntSlice(Engagement))
		i = 3
		total := 0
		for i < 12 {
			total = total + Engagement[i]
			i++
		}
		avgEngagement := float64(total) / (9 * float64(Followers))
		BestEngagement := (float64(TotalLikes) + float64(TotalComments)) / (12 * float64(Followers))

		MediaTimestamp := t - 1
		Days := 180
		i = 0
		for t-MediaTimestamp < (Days * 24 * 60 * 60) {
			if EndCursor != "" || len(EndCursor) > 0 {
				URL := "https://www.instagram.com/graphql/query/?query_hash=bfa387b2992c3a52dcbe447467b4b771&variables=%7B%22id%22%3A%22" + userId + "%22%2C%22first%22%3A12%2C%22after%22%3A%22" + EndCursor[:len(EndCursor)-2] + "%3D%3D%22%7D"
				fmt.Println(URL)
				resp, err := soup.Get(URL)
				if err != nil {
					fmt.Println("username not found")
				}
				fmt.Println(string(resp))
				var mediaResponse MediaResponse
				json.Unmarshal([]byte(resp), &mediaResponse)
				EndCursor = mediaResponse.Data.User.EdgeOwnerToTimelineMedia.PageInfo.EndCursor
				j := 0
				if len(mediaResponse.Data.User.EdgeOwnerToTimelineMedia.Edges) <= 0 {
					break
				}
				for j < len(mediaResponse.Data.User.EdgeOwnerToTimelineMedia.Edges) {
					MediaTimestamp = mediaResponse.Data.User.EdgeOwnerToTimelineMedia.Edges[j].Node.TakenAtTimestamp
					if t-MediaTimestamp < (30 * 24 * 60 * 60) {
						NumberOfPosts30Days++
					}
					if t-MediaTimestamp < (90 * 24 * 60 * 60) {
						NumberOfPosts90Days++
					}
					if t-MediaTimestamp < (180 * 24 * 60 * 60) {
						NumberOfPosts180Days++
					}
					j++
				}
				time.Sleep(1 * time.Second)
				i++
			} else {
				break
			}
		}
		row = append(row, float64(NumberOfPosts30Days)/4)
		row = append(row, float64(NumberOfPosts90Days)/12)
		row = append(row, float64(NumberOfPosts180Days)/24)
		row = append(row, BestEngagement)
		row = append(row, avgEngagement)
		row = append(row, (avgEngagement - BestEngagement))
		row = append(row, TotalComments/12)
		row = append(row, (BestEngagement)-(TotalLikes/(12*float64(Followers))))
		finalValues = append(finalValues, row)
		fmt.Println(finalValues)
	}
	return finalValues
}
func GetUserID(userName string) string {
	Url := "http://www.instagram.com/" + userName + "/"
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
	if err != nil {
		fmt.Println("Error")
		return ""
	} else {
		actual := strings.Index(string(body), "<script type=\"text/javascript\">window._sharedData")
		if actual == -1 {
			return ""
		}
		end := strings.Index(string(body), "<script type=\"text/javascript\">window.__initialDataLoaded(window._sharedData);</script>")
		if end == -1 {
			return ""
		}
		filteredString := (string(body)[actual+len("<script type=\"text/javascript\">window._sharedData")+2 : end-11])
		if filteredString == "" {
			return ""
		}
		var igResponse IGResponse
		json.Unmarshal([]byte(filteredString), &igResponse)
		if len(igResponse.EntryData.ProfilePage) > 0 {
			return igResponse.EntryData.ProfilePage[0].Graphql.User.ID
		}
	}
	return ""
}

func GetFollowers(userName string, MaxFollowers string) []string {
	MaxFollowersInt, err := strconv.Atoi(MaxFollowers)
	// MaxFollowersInt--
	if err != nil {
		MaxFollowersInt = 500
	}
	MaxFollowersCount := 0
	var finalValues []string
	var igFollowersResearch IGFollowersResearch
	EndCursor := "first"
	NextPage := true
	for NextPage && MaxFollowersCount < MaxFollowersInt {
		var URL string
		if EndCursor == "first" {
			URL = "https://www.instagram.com/graphql/query/?query_hash=d04b0a864b4b54837c0d870b0e77e076&variables=%7B%22id%22%3A%22" + GetUserID(userName) + "%22%2C%22include_reel%22%3Afalse%2C%22fetch_mutual%22%3Afalse%2C%22first%22%3A24%7D"
		} else {
			URL = "https://www.instagram.com/graphql/query/?query_hash=d04b0a864b4b54837c0d870b0e77e076&variables=%7B%22id%22%3A%22" + GetUserID(userName) + "%22%2C%22include_reel%22%3Afalse%2C%22fetch_mutual%22%3Afalse%2C%22first%22%3A12%2C%22after%22%3A%22" + EndCursor[:len(EndCursor)-2] + "%3D%3D%22%7D"
		}
		fmt.Println(URL)
		// resp, err := soup.Get(URL)
		// url := "https://www.instagram.com/graphql/query/?query_hash=d04b0a864b4b54837c0d870b0e77e076&variables=%257B%2522id%2522%253A%25222094200507%2522%252C%2522include_reel%2522%253Afalse%252C%2522fetch_mutual%2522%253Afalse%252C%2522first%2522%253A24%257D"
		method := "GET"

		client := &http.Client{}
		req, err := http.NewRequest(method, URL, nil)
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Add("accept", " */*")
		req.Header.Add("Cookie", "mid=XSMB8QAEAAEs3mQemNZLh2dhx98f; ig_did=EBB71BE2-8122-414C-9E28-4946DF598A00; datr=mgUcXzN0FK6UZc2wKHzFVdS8; fbm_124024574287414=base_domain=.instagram.com; shbid=18143; shbts=1599896664.3834596; rur=ATN; fbsr_124024574287414=Gx2jo4u1YNucR8uhdoS_OZz11ssN3ZH1Hm99OsmpsrE.eyJ1c2VyX2lkIjoiMTAwMDAyMDg2MzA5OTAzIiwiY29kZSI6IkFRREFOUlJ5VUtRN3BwaENSY1BGOVNLM3hvb1hVb2Q2UFBDMkJpaGtuUkVnd3BPQTZuc1dSVHdUbzNQUldxdjFnZi1nM0EwMlhDY01rZE9qQ1lzMzB1Z19KanZVUENVc2d5YzhFcml2cUNGU3pmOERaUUpRSXVFc2NxTGNNeWw0WlU2dURidHdZekRQWFJHQUhnQ0g0bkNvb3R0NnZyUWFkdGJ0SWV0d3BwcnZNc2hTbUJidmNab2tndkVxd3h1N3Jyd3FrU2F0OGdiT0xYWG1rV3p1T2QzT2tLUVRVdlBXc2xWOHpRellwal9sbjMzVjZPb0tFNmZMNm9TVnhNZk0wdU1aanprWDFPZ0IweFlmYU1PcHJGMW9qcUFmakJUMGJGUVl2LWVuZlNYeHIyS29uVS1LTzJrdl96TU9jVVNhTm5KeDJLVDN6NTcyMUhxenIxMlUtZDBRIiwib2F1dGhfdG9rZW4iOiJFQUFCd3pMaXhuallCQU0xbXMxQXJMMXQ2cU5rWGw5RTFnOXYyWkFyUHRyVk9wMHVFZmhHY05HMTJOaFpCZUhRNlFWalI5em1OcU85RU1Bc2pWOE85N2FXcVpCM2tQdkRaQncyb0gyelNrd3Azc0xRMXVMZ2p0OXlLOXE1WkN6QlBXaTRqdTl5bmtvT201NHB1d0s5MTFFSHdvSGgwZlZIWkNaQXkzSXdRa3hrT2NpbkRYZ0RzVzBaQiIsImFsZ29yaXRobSI6IkhNQUMtU0hBMjU2IiwiaXNzdWVkX2F0IjoxNTk5OTEwMzQwfQ; csrftoken=AYPfDg0kdLFbPNZbVHkkJIojQ1wPKSdH; ds_user_id=41309535897; sessionid=41309535897%3A4XfannYCtGdfzr%3A29;")

		// req.Header.Add("cookie", "mid=XSMB8QAEAAEs3mQemNZLh2dhx98f; ig_did=EBB71BE2-8122-414C-9E28-4946DF598A00; datr=mgUcXzN0FK6UZc2wKHzFVdS8; fbm_124024574287414=base_domain=.instagram.com; shbid=18143; shbts=1599896664.3834596; rur=ATN; fbsr_124024574287414=gO0ZKWcK57Alnu3dYwqdbq0dRLKS9Dkj6l1TY1m4HoU.eyJ1c2VyX2lkIjoiMTAwMDAyMDg2MzA5OTAzIiwiY29kZSI6IkFRQ3ZUSlVvTkFxM2Nqbzc2TGZta3lGa3pFel9Td0pjS2hpOVFBSTBNaUs3aUdzTXYyWUU1Tkp3MGkyN0Z1UFU2N2pCM2wtVVRubjdlMkhkQTVaZ1ZWQjZZWDVuVGNvNE5HQ3VHZklJczZwTW1sdkh2ZXVnY1V6ZHZGajZqWW1TMUlvX2lMdDB1c3V2emg3U2ZNUkxCY3FMeEhVTnVQS3pIZkc2VTJ1bHhRV1EwQVVVYkNOVkduQ0dVQy1jUlZWRFVEMGVUckxEb0VNSDBTNGtKODk0bk9Valh0WnhMOEVsTWxGTmdZVHhVZmtRUUEwcnhfNnRfUlltdlJTVnZzb2hzRE1WVVNfOXVrbE9Rc1djM0JEdm5PeFR4WDhBaXhkMERsN25zOW9SeFNmczRWeU1vNldTYWFkWTBhaElpSHVRNjh2Z0dqQ2xfNmlxNEZmVmNlb2M1Mk81Iiwib2F1dGhfdG9rZW4iOiJFQUFCd3pMaXhuallCQUIxY252YVNqUmtuY3kxeUNEUnM2cTVBWkF3N0ozYWRlaWRhSHlLVkxwY1pBUzRScHBmT3lPR1F1Nm03SEtwOUJCRzZQcFhkdjhaQ3BWd1pBT0ZaQWxXU0p5WGdSOFkxTExDT0prS2NYbjhpT1l1MDUzc2lPcHJja3U4SXd3YWdVb1pCcmFraldQTzlaQzhIUHNqMkZMZFZJMEF6TWNzQ1haQjlucG5jNFk5byIsImFsZ29yaXRobSI6IkhNQUMtU0hBMjU2IiwiaXNzdWVkX2F0IjoxNTk5OTAwMjIxfQ; csrftoken=CZVniHggvmGG9S2FIUlnTmzzACw0hWAF; sessionid=1270182093%3AEP9oNwncijl685%3A8;")

		res, err := client.Do(req)
		if err != nil {
			fmt.Println("username not found")
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("username not found")
		}
		fmt.Println(string(body))
		json.Unmarshal([]byte(body), &igFollowersResearch)
		iterator := 0
		for iterator < len(igFollowersResearch.Data.User.EdgeFollow.Edges) {
			if igFollowersResearch.Data.User.EdgeFollow.Edges[iterator].Node.Username != "" && MaxFollowersCount < MaxFollowersInt {
				MaxFollowersCount++
				finalValues = append(finalValues, igFollowersResearch.Data.User.EdgeFollow.Edges[iterator].Node.Username)
				iterator++
			} else {
				break
			}
		}
		EndCursor = igFollowersResearch.Data.User.EdgeFollow.PageInfo.EndCursor
		NextPage = igFollowersResearch.Data.User.EdgeFollow.PageInfo.HasNextPage
	}
	fmt.Println(finalValues)
	return finalValues
}

func GetIGReport(userNames []string, SearchQuery map[string]int) [][]interface{} {
	var finalValues [][]interface{}
	parentIterator := 0
	for parentIterator < len(userNames) {
		var row []interface{}
		time.Sleep(100 * time.Millisecond)
		Url := "http://www.instagram.com/" + userNames[parentIterator] + "/"
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
		if err != nil {
			fmt.Println("Error")
			continue
		} else {
			actual := strings.Index(string(body), "<script type=\"text/javascript\">window._sharedData")
			end := strings.Index(string(body), "<script type=\"text/javascript\">window.__initialDataLoaded(window._sharedData);</script>")
			filteredString := (string(body)[actual+len("<script type=\"text/javascript\">window._sharedData")+2 : end-11])
			fmt.Println(filteredString)
			var igResponse IGResponse
			json.Unmarshal([]byte(filteredString), &igResponse)
			TotalLikes := 0.0
			TotalComments := 0.0
			if len(igResponse.EntryData.ProfilePage) > 0 {
				Followers := igResponse.EntryData.ProfilePage[0].Graphql.User.EdgeFollowedBy.Count
				if val, ok := SearchQuery["MinFollower"]; ok {
					if Followers < val {
						parentIterator++
						continue
					}
				}
				if val, ok := SearchQuery["MaxFollower"]; ok {
					if Followers > val {
						parentIterator++
						continue
					}
				}

				i := 0
				Engagement := make([]int, 12)
				for i < 12 && i < len(igResponse.EntryData.ProfilePage[0].Graphql.User.EdgeOwnerToTimelineMedia.Edges) {
					Likes := igResponse.EntryData.ProfilePage[0].Graphql.User.EdgeOwnerToTimelineMedia.Edges[i].Node.EdgeLikedBy.Count
					Comments := igResponse.EntryData.ProfilePage[0].Graphql.User.EdgeOwnerToTimelineMedia.Edges[i].Node.EdgeMediaToComment.Count
					TotalLikes = TotalLikes + float64(Likes)
					TotalComments = TotalComments + float64(Comments)
					Engagement[i] = (Likes + Comments)
					i++
				}
				sort.Sort(sort.IntSlice(Engagement))
				i = 3
				total := 0
				for i < 12 {
					total = total + Engagement[i]
					i++
				}
				BestEngagement := float64(total) / (9 * float64(Followers))
				BestEngagement = BestEngagement * 100
				avgEngagement := (float64(TotalLikes) + float64(TotalComments)) / (12 * float64(Followers))
				avgEngagement = avgEngagement * 100
				if val, ok := SearchQuery["MinN"]; ok {
					avgEngagementInt := int(math.Round(avgEngagement))
					if avgEngagementInt < val {
						parentIterator++
						continue
					}
				}
				if val, ok := SearchQuery["MinNStar"]; ok {
					BestEngagementInt := int(math.Round(BestEngagement))
					if BestEngagementInt < val {
						parentIterator++
						continue
					}
				}
				row = append(row, userNames[parentIterator])
			}
		}
		finalValues = append(finalValues, row)
		parentIterator++
	}
	return finalValues
}

func GetFinalMapOfResearchedUsers() {

}

type IGResponse struct {
	EntryData struct {
		ProfilePage []struct {
			LoggingPageID         string `json:"logging_page_id"`
			ShowSuggestedProfiles bool   `json:"show_suggested_profiles"`
			ShowFollowDialog      bool   `json:"show_follow_dialog"`
			Graphql               struct {
				User struct {
					Biography              string      `json:"biography"`
					BlockedByViewer        bool        `json:"blocked_by_viewer"`
					BusinessEmail          interface{} `json:"business_email"`
					RestrictedByViewer     interface{} `json:"restricted_by_viewer"`
					CountryBlock           bool        `json:"country_block"`
					ExternalURL            string      `json:"external_url"`
					ExternalURLLinkshimmed string      `json:"external_url_linkshimmed"`
					EdgeFollowedBy         struct {
						Count int `json:"count"`
					} `json:"edge_followed_by"`
					FollowedByViewer bool `json:"followed_by_viewer"`
					EdgeFollow       struct {
						Count int `json:"count"`
					} `json:"edge_follow"`
					FollowsViewer        bool        `json:"follows_viewer"`
					FullName             string      `json:"full_name"`
					HasArEffects         bool        `json:"has_ar_effects"`
					HasGuides            bool        `json:"has_guides"`
					HasChannel           bool        `json:"has_channel"`
					HasBlockedViewer     bool        `json:"has_blocked_viewer"`
					HighlightReelCount   int         `json:"highlight_reel_count"`
					HasRequestedViewer   bool        `json:"has_requested_viewer"`
					ID                   string      `json:"id"`
					IsBusinessAccount    bool        `json:"is_business_account"`
					IsJoinedRecently     bool        `json:"is_joined_recently"`
					BusinessCategoryName interface{} `json:"business_category_name"`
					OverallCategoryName  interface{} `json:"overall_category_name"`
					CategoryEnum         interface{} `json:"category_enum"`
					IsPrivate            bool        `json:"is_private"`
					IsVerified           bool        `json:"is_verified"`
					EdgeMutualFollowedBy struct {
						Count int           `json:"count"`
						Edges []interface{} `json:"edges"`
					} `json:"edge_mutual_followed_by"`
					ProfilePicURL          string      `json:"profile_pic_url"`
					ProfilePicURLHd        string      `json:"profile_pic_url_hd"`
					RequestedByViewer      bool        `json:"requested_by_viewer"`
					Username               string      `json:"username"`
					ConnectedFbPage        interface{} `json:"connected_fb_page"`
					EdgeFelixVideoTimeline struct {
						Count    int `json:"count"`
						PageInfo struct {
							HasNextPage bool        `json:"has_next_page"`
							EndCursor   interface{} `json:"end_cursor"`
						} `json:"page_info"`
						Edges []struct {
							Node struct {
								Typename   string `json:"__typename"`
								ID         string `json:"id"`
								Shortcode  string `json:"shortcode"`
								Dimensions struct {
									Height int `json:"height"`
									Width  int `json:"width"`
								} `json:"dimensions"`
								DisplayURL            string `json:"display_url"`
								EdgeMediaToTaggedUser struct {
									Edges []interface{} `json:"edges"`
								} `json:"edge_media_to_tagged_user"`
								FactCheckOverallRating interface{} `json:"fact_check_overall_rating"`
								FactCheckInformation   interface{} `json:"fact_check_information"`
								GatingInfo             interface{} `json:"gating_info"`
								MediaOverlayInfo       interface{} `json:"media_overlay_info"`
								MediaPreview           string      `json:"media_preview"`
								Owner                  struct {
									ID       string `json:"id"`
									Username string `json:"username"`
								} `json:"owner"`
								IsVideo              bool        `json:"is_video"`
								AccessibilityCaption interface{} `json:"accessibility_caption"`
								DashInfo             struct {
									IsDashEligible    bool        `json:"is_dash_eligible"`
									VideoDashManifest interface{} `json:"video_dash_manifest"`
									NumberOfQualities int         `json:"number_of_qualities"`
								} `json:"dash_info"`
								HasAudio           bool   `json:"has_audio"`
								TrackingToken      string `json:"tracking_token"`
								VideoURL           string `json:"video_url"`
								VideoViewCount     int    `json:"video_view_count"`
								EdgeMediaToCaption struct {
									Edges []struct {
										Node struct {
											Text string `json:"text"`
										} `json:"node"`
									} `json:"edges"`
								} `json:"edge_media_to_caption"`
								EdgeMediaToComment struct {
									Count int `json:"count"`
								} `json:"edge_media_to_comment"`
								CommentsDisabled bool `json:"comments_disabled"`
								TakenAtTimestamp int  `json:"taken_at_timestamp"`
								EdgeLikedBy      struct {
									Count int `json:"count"`
								} `json:"edge_liked_by"`
								EdgeMediaPreviewLike struct {
									Count int `json:"count"`
								} `json:"edge_media_preview_like"`
								Location           interface{} `json:"location"`
								ThumbnailSrc       string      `json:"thumbnail_src"`
								ThumbnailResources []struct {
									Src          string `json:"src"`
									ConfigWidth  int    `json:"config_width"`
									ConfigHeight int    `json:"config_height"`
								} `json:"thumbnail_resources"`
								FelixProfileGridCrop interface{} `json:"felix_profile_grid_crop"`
								EncodingStatus       interface{} `json:"encoding_status"`
								IsPublished          bool        `json:"is_published"`
								ProductType          string      `json:"product_type"`
								Title                string      `json:"title"`
								VideoDuration        float64     `json:"video_duration"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"edge_felix_video_timeline"`
					EdgeOwnerToTimelineMedia struct {
						Count    int `json:"count"`
						PageInfo struct {
							HasNextPage bool   `json:"has_next_page"`
							EndCursor   string `json:"end_cursor"`
						} `json:"page_info"`
						Edges []struct {
							Node struct {
								Typename   string `json:"__typename"`
								ID         string `json:"id"`
								Shortcode  string `json:"shortcode"`
								Dimensions struct {
									Height int `json:"height"`
									Width  int `json:"width"`
								} `json:"dimensions"`
								DisplayURL            string `json:"display_url"`
								EdgeMediaToTaggedUser struct {
									Edges []interface{} `json:"edges"`
								} `json:"edge_media_to_tagged_user"`
								FactCheckOverallRating interface{} `json:"fact_check_overall_rating"`
								FactCheckInformation   interface{} `json:"fact_check_information"`
								GatingInfo             interface{} `json:"gating_info"`
								MediaOverlayInfo       interface{} `json:"media_overlay_info"`
								MediaPreview           string      `json:"media_preview"`
								Owner                  struct {
									ID       string `json:"id"`
									Username string `json:"username"`
								} `json:"owner"`
								IsVideo              bool        `json:"is_video"`
								AccessibilityCaption interface{} `json:"accessibility_caption"`
								DashInfo             struct {
									IsDashEligible    bool        `json:"is_dash_eligible"`
									VideoDashManifest interface{} `json:"video_dash_manifest"`
									NumberOfQualities int         `json:"number_of_qualities"`
								} `json:"dash_info"`
								HasAudio           bool   `json:"has_audio"`
								TrackingToken      string `json:"tracking_token"`
								VideoURL           string `json:"video_url"`
								VideoViewCount     int    `json:"video_view_count"`
								EdgeMediaToCaption struct {
									Edges []struct {
										Node struct {
											Text string `json:"text"`
										} `json:"node"`
									} `json:"edges"`
								} `json:"edge_media_to_caption"`
								EdgeMediaToComment struct {
									Count int `json:"count"`
								} `json:"edge_media_to_comment"`
								CommentsDisabled bool `json:"comments_disabled"`
								TakenAtTimestamp int  `json:"taken_at_timestamp"`
								EdgeLikedBy      struct {
									Count int `json:"count"`
								} `json:"edge_liked_by"`
								EdgeMediaPreviewLike struct {
									Count int `json:"count"`
								} `json:"edge_media_preview_like"`
								Location           interface{} `json:"location"`
								ThumbnailSrc       string      `json:"thumbnail_src"`
								ThumbnailResources []struct {
									Src          string `json:"src"`
									ConfigWidth  int    `json:"config_width"`
									ConfigHeight int    `json:"config_height"`
								} `json:"thumbnail_resources"`
								FelixProfileGridCrop interface{} `json:"felix_profile_grid_crop"`
								ProductType          string      `json:"product_type"`
							} `json:"node,omitempty"`
						} `json:"edges"`
					} `json:"edge_owner_to_timeline_media"`
					EdgeSavedMedia struct {
						Count    int `json:"count"`
						PageInfo struct {
							HasNextPage bool        `json:"has_next_page"`
							EndCursor   interface{} `json:"end_cursor"`
						} `json:"page_info"`
						Edges []interface{} `json:"edges"`
					} `json:"edge_saved_media"`
					EdgeMediaCollections struct {
						Count    int `json:"count"`
						PageInfo struct {
							HasNextPage bool        `json:"has_next_page"`
							EndCursor   interface{} `json:"end_cursor"`
						} `json:"page_info"`
						Edges []interface{} `json:"edges"`
					} `json:"edge_media_collections"`
					EdgeRelatedProfiles struct {
						Edges []struct {
							Node struct {
								ID            string `json:"id"`
								FullName      string `json:"full_name"`
								IsPrivate     bool   `json:"is_private"`
								IsVerified    bool   `json:"is_verified"`
								ProfilePicURL string `json:"profile_pic_url"`
								Username      string `json:"username"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"edge_related_profiles"`
				} `json:"user"`
			} `json:"graphql"`
			ToastContentOnLoad interface{} `json:"toast_content_on_load"`
		} `json:"ProfilePage"`
	} `json:"entry_data"`
	Hostname              string  `json:"hostname"`
	IsWhitelistedCrawlBot bool    `json:"is_whitelisted_crawl_bot"`
	DeploymentStage       string  `json:"deployment_stage"`
	Platform              string  `json:"platform"`
	Nonce                 string  `json:"nonce"`
	MidPct                float64 `json:"mid_pct"`
	ZeroData              struct {
	} `json:"zero_data"`
	CacheSchemaVersion int `json:"cache_schema_version"`
	ServerChecks       struct {
	} `json:"server_checks"`
	DeviceID          string `json:"device_id"`
	BrowserPushPubKey string `json:"browser_push_pub_key"`
	Encryption        struct {
		KeyID     string `json:"key_id"`
		PublicKey string `json:"public_key"`
		Version   string `json:"version"`
	} `json:"encryption"`
	IsDev                  bool        `json:"is_dev"`
	SignalCollectionConfig interface{} `json:"signal_collection_config"`
	RolloutHash            string      `json:"rollout_hash"`
	BundleVariant          string      `json:"bundle_variant"`
	FrontendEnv            string      `json:"frontend_env"`
}

type MediaResponse struct {
	Data struct {
		User struct {
			EdgeOwnerToTimelineMedia struct {
				Count    int `json:"count"`
				PageInfo struct {
					HasNextPage bool   `json:"has_next_page"`
					EndCursor   string `json:"end_cursor"`
				} `json:"page_info"`
				Edges []struct {
					Node struct {
						Typename                string      `json:"__typename"`
						ID                      string      `json:"id"`
						GatingInfo              interface{} `json:"gating_info"`
						FactCheckOverallRating  interface{} `json:"fact_check_overall_rating"`
						FactCheckInformation    interface{} `json:"fact_check_information"`
						MediaOverlayInfo        interface{} `json:"media_overlay_info"`
						SensitivityFrictionInfo interface{} `json:"sensitivity_friction_info"`
						Dimensions              struct {
							Height int `json:"height"`
							Width  int `json:"width"`
						} `json:"dimensions"`
						DisplayURL       string `json:"display_url"`
						DisplayResources []struct {
							Src          string `json:"src"`
							ConfigWidth  int    `json:"config_width"`
							ConfigHeight int    `json:"config_height"`
						} `json:"display_resources"`
						IsVideo               bool        `json:"is_video"`
						MediaPreview          interface{} `json:"media_preview"`
						TrackingToken         string      `json:"tracking_token"`
						EdgeMediaToTaggedUser struct {
							Edges []interface{} `json:"edges"`
						} `json:"edge_media_to_tagged_user"`
						AccessibilityCaption interface{} `json:"accessibility_caption"`
						EdgeMediaToCaption   struct {
							Edges []struct {
								Node struct {
									Text string `json:"text"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"edge_media_to_caption"`
						Shortcode          string `json:"shortcode"`
						EdgeMediaToComment struct {
							Count    int `json:"count"`
							PageInfo struct {
								HasNextPage bool   `json:"has_next_page"`
								EndCursor   string `json:"end_cursor"`
							} `json:"page_info"`
							Edges []struct {
								Node struct {
									ID              string `json:"id"`
									Text            string `json:"text"`
									CreatedAt       int    `json:"created_at"`
									DidReportAsSpam bool   `json:"did_report_as_spam"`
									Owner           struct {
										ID            string `json:"id"`
										IsVerified    bool   `json:"is_verified"`
										ProfilePicURL string `json:"profile_pic_url"`
										Username      string `json:"username"`
									} `json:"owner"`
									ViewerHasLiked bool `json:"viewer_has_liked"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"edge_media_to_comment"`
						EdgeMediaToSponsorUser struct {
							Edges []interface{} `json:"edges"`
						} `json:"edge_media_to_sponsor_user"`
						CommentsDisabled     bool `json:"comments_disabled"`
						TakenAtTimestamp     int  `json:"taken_at_timestamp"`
						EdgeMediaPreviewLike struct {
							Count int           `json:"count"`
							Edges []interface{} `json:"edges"`
						} `json:"edge_media_preview_like"`
						Owner struct {
							ID       string `json:"id"`
							Username string `json:"username"`
						} `json:"owner"`
						Location                   interface{} `json:"location"`
						ViewerHasLiked             bool        `json:"viewer_has_liked"`
						ViewerHasSaved             bool        `json:"viewer_has_saved"`
						ViewerHasSavedToCollection bool        `json:"viewer_has_saved_to_collection"`
						ViewerInPhotoOfYou         bool        `json:"viewer_in_photo_of_you"`
						ViewerCanReshare           bool        `json:"viewer_can_reshare"`
						ThumbnailSrc               string      `json:"thumbnail_src"`
						ThumbnailResources         []struct {
							Src          string `json:"src"`
							ConfigWidth  int    `json:"config_width"`
							ConfigHeight int    `json:"config_height"`
						} `json:"thumbnail_resources"`
						EdgeSidecarToChildren struct {
							Edges []struct {
								Node struct {
									Typename                string      `json:"__typename"`
									ID                      string      `json:"id"`
									GatingInfo              interface{} `json:"gating_info"`
									FactCheckOverallRating  interface{} `json:"fact_check_overall_rating"`
									FactCheckInformation    interface{} `json:"fact_check_information"`
									MediaOverlayInfo        interface{} `json:"media_overlay_info"`
									SensitivityFrictionInfo interface{} `json:"sensitivity_friction_info"`
									Dimensions              struct {
										Height int `json:"height"`
										Width  int `json:"width"`
									} `json:"dimensions"`
									DisplayURL       string `json:"display_url"`
									DisplayResources []struct {
										Src          string `json:"src"`
										ConfigWidth  int    `json:"config_width"`
										ConfigHeight int    `json:"config_height"`
									} `json:"display_resources"`
									IsVideo               bool   `json:"is_video"`
									MediaPreview          string `json:"media_preview"`
									TrackingToken         string `json:"tracking_token"`
									EdgeMediaToTaggedUser struct {
										Edges []interface{} `json:"edges"`
									} `json:"edge_media_to_tagged_user"`
									AccessibilityCaption interface{} `json:"accessibility_caption"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"edge_sidecar_to_children"`
					} `json:"node,omitempty"`
				} `json:"edges"`
			} `json:"edge_owner_to_timeline_media"`
		} `json:"user"`
	} `json:"data"`
	Status string `json:"status"`
}

type IGFollowersResearch struct {
	Data struct {
		User struct {
			EdgeFollow struct {
				Count    int `json:"count"`
				PageInfo struct {
					HasNextPage bool   `json:"has_next_page"`
					EndCursor   string `json:"end_cursor"`
				} `json:"page_info"`
				Edges []struct {
					Node struct {
						ID                string `json:"id"`
						Username          string `json:"username"`
						FullName          string `json:"full_name"`
						ProfilePicURL     string `json:"profile_pic_url"`
						IsPrivate         bool   `json:"is_private"`
						IsVerified        bool   `json:"is_verified"`
						FollowedByViewer  bool   `json:"followed_by_viewer"`
						RequestedByViewer bool   `json:"requested_by_viewer"`
						Reel              struct {
							ID              string      `json:"id"`
							ExpiringAt      int         `json:"expiring_at"`
							HasPrideMedia   bool        `json:"has_pride_media"`
							LatestReelMedia int         `json:"latest_reel_media"`
							Seen            interface{} `json:"seen"`
							Owner           struct {
								Typename      string `json:"__typename"`
								ID            string `json:"id"`
								ProfilePicURL string `json:"profile_pic_url"`
								Username      string `json:"username"`
							} `json:"owner"`
						} `json:"reel"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"edge_follow"`
		} `json:"user"`
	} `json:"data"`
	Status string `json:"status"`
}
