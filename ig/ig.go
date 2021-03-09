package ig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
	"github.com/bhambri94/ig-reports/configs"
)

func GetAccountFollowersDetails(userName string, MaxFollowers string, SessionID string) ([][]interface{}, string) {
	var finalValues [][]interface{}
	if SessionID == "" {
		SessionID = configs.Configurations.SessionId
	}
	MaxFollowersInt, err := strconv.Atoi(MaxFollowers)
	if err != nil {
		MaxFollowersInt = 500
	}
	UserID, _, CookieErrorString := GetUserIDAndFollower(userName, GetRandomCookie(SessionID))

	fmt.Println("*****")
	fmt.Println("User Id found: " + UserID)
	fmt.Println("*****")
	if UserID == "" {
		return nil, CookieErrorString
	}
	var igFollowersResearch IGFollowersResearch
	EndCursor := "first"
	NextPage := true
	MaxFollowersCount := 0
	for NextPage && MaxFollowersCount < MaxFollowersInt {
		var URL string
		if EndCursor == "first" {
			URL = "https://www.instagram.com/graphql/query/?query_hash=d04b0a864b4b54837c0d870b0e77e076&variables=%7B%22id%22%3A%22" + UserID + "%22%2C%22include_reel%22%3Afalse%2C%22fetch_mutual%22%3Afalse%2C%22first%22%3A24%7D"
		} else {
			URL = "https://www.instagram.com/graphql/query/?query_hash=d04b0a864b4b54837c0d870b0e77e076&variables=%7B%22id%22%3A%22" + UserID + "%22%2C%22include_reel%22%3Afalse%2C%22fetch_mutual%22%3Afalse%2C%22first%22%3A12%2C%22after%22%3A%22" + EndCursor[:len(EndCursor)-2] + "%3D%3D%22%7D"
		}
		fmt.Println(URL)
		method := "GET"

		client := &http.Client{}
		req, err := http.NewRequest(method, URL, nil)
		if err != nil {
			fmt.Println(err)
		}
		if SessionID == "" {
			SessionID = configs.Configurations.SessionId
		}
		req.Header.Add("accept", " */*")
		req.Header.Add("Cookie", "sessionid="+GetRandomCookie(SessionID))

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
			var row []interface{}
			if igFollowersResearch.Data.User.EdgeFollow.Edges[iterator].Node.Username != "" && MaxFollowersCount < MaxFollowersInt {
				MaxFollowersCount++
				row = append(row, igFollowersResearch.Data.User.EdgeFollow.Edges[iterator].Node.FullName, igFollowersResearch.Data.User.EdgeFollow.Edges[iterator].Node.Username)
				iterator++
			} else {
				break
			}
			finalValues = append(finalValues, row)
		}
		EndCursor = igFollowersResearch.Data.User.EdgeFollow.PageInfo.EndCursor
		NextPage = igFollowersResearch.Data.User.EdgeFollow.PageInfo.HasNextPage
	}
	return finalValues, CookieErrorString
}

func GetLatestFollowingCount(userId string, SessionID string) string {
	req, err := http.NewRequest("GET", "https://www.instagram.com/graphql/query/?query_hash=3dec7e2c57367ef3da3d987d89f9dbc8&variables=%7B%22id%22%3A%22"+userId+"%22%2C%22include_reel%22%3Atrue%2C%22fetch_mutual%22%3Afalse%2C%22first%22%3A12%7D", nil)
	if err != nil {
		// handle err
	}
	Cookie := GetRandomCookie(SessionID)
	req.Header.Set("Authority", "www.instagram.com")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Add("Cookie", "sessionid="+Cookie)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("username not found")
	}
	var igFollowingAllResearch IGFollowingAllResearch
	json.Unmarshal([]byte(body), &igFollowingAllResearch)
	LatestFollowingCount := igFollowingAllResearch.Data.User.EdgeFollow.Count
	return strconv.Itoa(LatestFollowingCount)
}

func GetReportNew(userName string, SessionID string) ([][]interface{}, string) {
	var finalValues [][]interface{}
	NumberOfPosts30Days := 0
	NumberOfPosts90Days := 0
	NumberOfPosts180Days := 0
	if SessionID == "" {
		SessionID = configs.Configurations.SessionId
	}
	UserId, _, CookieErrorString := GetUserIDAndFollower(userName, GetRandomCookie(SessionID))
	Url := "https://www.instagram.com/graphql/query/?query_hash=bfa387b2992c3a52dcbe447467b4b771&variables=%7B%22id%22%3A%22" + UserId + "%22%2C%22first%22%3A12%7D"
	fmt.Println(Url)
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, Url, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("accept", " */*")
	if SessionID == "" {
		SessionID = configs.Configurations.SessionId
	}
	Cookie := GetRandomCookie(SessionID)
	req.Header.Add("Cookie", "sessionid="+Cookie)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("username not found")
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("username not found")
	}
	fmt.Println(body)
	var igResponse AutoGenerated
	json.Unmarshal([]byte(body), &igResponse)
	loc, _ := time.LoadLocation("Europe/Rome")
	currentTime := time.Now().In(loc)
	Time := currentTime.Format("2006-01-02")
	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	timestamp = timestamp[:10]
	t, _ := strconv.Atoi(timestamp)
	var row []interface{}
	TotalLikes := 0.0
	TotalComments := 0.0
	StandardDeviation := 0.0
	Variance := 0.0
	NumberOfPostsOnFirstPage := 12
	FirstPage := true
	if len(igResponse.Data.User.EdgeOwnerToTimelineMedia.Edges) > 0 {
		FollowerURL := "https://www.instagram.com/graphql/query/?query_hash=c76146de99bb02f6415203be841dd25a&variables=%7B%22id%22%3A%22" + UserId + "%22%2C%22include_reel%22%3Afalse%2C%22fetch_mutual%22%3Afalse%2C%22first%22%3A24%7D"
		fmt.Println(FollowerURL)
		method := "GET"
		client := &http.Client{}
		req, err := http.NewRequest(method, FollowerURL, nil)
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Add("accept", " */*")
		if SessionID == "" {
			SessionID = configs.Configurations.SessionId
		}
		Cookie := GetRandomCookie(SessionID)
		req.Header.Add("Cookie", "sessionid="+Cookie)
		res, err := client.Do(req)
		if err != nil {
			fmt.Println("username not found")
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("username not found")
		}
		var igFollowersResearch IGFollowersAllResearch
		json.Unmarshal([]byte(body), &igFollowersResearch)
		userId := UserId
		EndCursor := igResponse.Data.User.EdgeOwnerToTimelineMedia.PageInfo.EndCursor
		row = append(row, Time)
		row = append(row, userName)
		Followers := igFollowersResearch.Data.User.EdgeFollowedBy.Count
		row = append(row, Followers)
		row = append(row, igResponse.Data.User.EdgeOwnerToTimelineMedia.Count)
		if FirstPage {
			NumberOfPostsOnFirstPage = len(igResponse.Data.User.EdgeOwnerToTimelineMedia.Edges)
			FirstPage = false
		}
		i := 0
		Engagement := make([]int, 12)
		for i < 12 && i < len(igResponse.Data.User.EdgeOwnerToTimelineMedia.Edges) {
			Likes := igResponse.Data.User.EdgeOwnerToTimelineMedia.Edges[i].Node.EdgeMediaPreviewLike.Count
			Comments := igResponse.Data.User.EdgeOwnerToTimelineMedia.Edges[i].Node.EdgeMediaToComment.Count
			TotalLikes = TotalLikes + float64(Likes)
			TotalComments = TotalComments + float64(Comments)
			Engagement[i] = (Likes + Comments)
			MediaTimestamp := igResponse.Data.User.EdgeOwnerToTimelineMedia.Edges[i].Node.TakenAtTimestamp
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
		fmt.Println("Follower Count is :" + strconv.Itoa(Followers))
		if Followers == 0 {
			return nil, CookieErrorString
		}
		avgEngagement := float64(total) / (9 * float64(Followers))
		BestEngagement := (float64(TotalLikes) + float64(TotalComments)) / (float64(NumberOfPostsOnFirstPage) * float64(Followers))

		i = 0
		var sd float64
		for i < 12 {
			sd += math.Pow(((float64(Engagement[i]) / float64(Followers)) - BestEngagement), 2)
			i++
		}
		Variance = sd / (float64(NumberOfPostsOnFirstPage))
		StandardDeviation = math.Sqrt(Variance)
		MediaTimestamp := t - 1
		Days := 90
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
		// row = append(row, "NA")
		row = append(row, BestEngagement)
		row = append(row, avgEngagement)
		row = append(row, (avgEngagement - BestEngagement))
		row = append(row, Variance)
		row = append(row, StandardDeviation)
		row = append(row, TotalComments/float64(NumberOfPostsOnFirstPage))
		row = append(row, (BestEngagement)-(TotalLikes/(float64(NumberOfPostsOnFirstPage)*float64(Followers))))
		finalValues = append(finalValues, row)
	}
	return finalValues, CookieErrorString
}

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
	}
	return finalValues
}

func GetUserIDAndFollowerFromCodeNinja(userName string) (string, int) {

	req, err := http.NewRequest("GET", "https://www.instagram.com/web/search/topsearch/?query="+userName, nil)
	if err != nil {
		// handle err
	}
	req.Header.Set("Authority", "www.instagram.com")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", "https://codeofaninja.com")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Referer", "https://codeofaninja.com/tools/find-instagram-user-id/")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.9")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	var instaID AutoGeneratedInstaID
	json.Unmarshal(body, &instaID)
	if len(instaID.Users) > 0 {
		return instaID.Users[0].User.Pk, 0
	} else {
		return "", 0
	}
}

type AutoGeneratedInstaID struct {
	Users []struct {
		Position int `json:"position"`
		User     struct {
			Pk                         string        `json:"pk"`
			Username                   string        `json:"username"`
			FullName                   string        `json:"full_name"`
			IsPrivate                  bool          `json:"is_private"`
			ProfilePicURL              string        `json:"profile_pic_url"`
			ProfilePicID               string        `json:"profile_pic_id"`
			IsVerified                 bool          `json:"is_verified"`
			HasAnonymousProfilePicture bool          `json:"has_anonymous_profile_picture"`
			MutualFollowersCount       int           `json:"mutual_followers_count"`
			AccountBadges              []interface{} `json:"account_badges"`
			LatestReelMedia            int           `json:"latest_reel_media"`
		} `json:"user"`
	} `json:"users"`
	Places []struct {
		Place struct {
			Location struct {
				Pk               string `json:"pk"`
				Name             string `json:"name"`
				Address          string `json:"address"`
				City             string `json:"city"`
				ShortName        string `json:"short_name"`
				ExternalSource   string `json:"external_source"`
				FacebookPlacesID int64  `json:"facebook_places_id"`
			} `json:"location"`
			Title        string        `json:"title"`
			Subtitle     string        `json:"subtitle"`
			MediaBundles []interface{} `json:"media_bundles"`
			Slug         string        `json:"slug"`
		} `json:"place"`
		Position int `json:"position"`
	} `json:"places"`
	Hashtags         []interface{} `json:"hashtags"`
	HasMore          bool          `json:"has_more"`
	RankToken        string        `json:"rank_token"`
	ClearClientCache bool          `json:"clear_client_cache"`
	Status           string        `json:"status"`
}

type InstaID struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	FullName      string `json:"full_name"`
	ProfilePicURL string `json:"profile_pic_url"`
	Followers     int    `json:"followers"`
	Followed      int    `json:"followed"`
	Biography     string `json:"biography"`
}

type NewInstaID struct {
	Users []struct {
		Position int `json:"position"`
		User     struct {
			Pk                         string        `json:"pk"`
			Username                   string        `json:"username"`
			FullName                   string        `json:"full_name"`
			IsPrivate                  bool          `json:"is_private"`
			ProfilePicURL              string        `json:"profile_pic_url"`
			ProfilePicID               string        `json:"profile_pic_id"`
			IsVerified                 bool          `json:"is_verified"`
			HasAnonymousProfilePicture bool          `json:"has_anonymous_profile_picture"`
			MutualFollowersCount       int           `json:"mutual_followers_count"`
			AccountBadges              []interface{} `json:"account_badges"`
			FriendshipStatus           struct {
				Following       bool `json:"following"`
				IsPrivate       bool `json:"is_private"`
				IncomingRequest bool `json:"incoming_request"`
				OutgoingRequest bool `json:"outgoing_request"`
				IsBestie        bool `json:"is_bestie"`
				IsRestricted    bool `json:"is_restricted"`
			} `json:"friendship_status"`
			LatestReelMedia int `json:"latest_reel_media"`
		} `json:"user,omitempty"`
		Places   []interface{} `json:"places"`
		Hashtags []struct {
			Position int `json:"position"`
			Hashtag  struct {
				Name                 string `json:"name"`
				ID                   int64  `json:"id"`
				MediaCount           int    `json:"media_count"`
				UseDefaultAvatar     bool   `json:"use_default_avatar"`
				ProfilePicURL        string `json:"profile_pic_url"`
				SearchResultSubtitle string `json:"search_result_subtitle"`
			} `json:"hashtag"`
		} `json:"hashtags"`
		HasMore          bool        `json:"has_more"`
		RankToken        string      `json:"rank_token"`
		ClearClientCache interface{} `json:"clear_client_cache"`
		Status           string      `json:"status"`
	}
}

func GetUserIDAndFollower(userName string, SessionID string) (string, int, string) {
	Url := "https://www.instagram.com/web/search/topsearch/?context=blended&query=" + userName + "&rank_token=0.7584840628946048"
	req, err := http.NewRequest("GET", Url, nil)
	if err != nil {
		// handle err
	}
	req.Header.Set("Authority", "www.instagram.com")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 9; SM-A102U Build/PPR1.180610.011; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/74.0.3729.136 Mobile Safari/537.36 Instagram 155.0.0.37.107 Android (28/9; 320dpi; 720x1468; samsung; SM-A102U; a10e; exynos7885; en_US; 239490550)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Add("Cookie", "sessionid="+SessionID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, "Issue with Cookie:" + SessionID
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	var newInstaID NewInstaID
	json.Unmarshal([]byte(body), &newInstaID)
	if len(newInstaID.Users) > 0 {
		if len(newInstaID.Users[0].User.Pk) > 0 {
			return newInstaID.Users[0].User.Pk, 0, ""
		}
	}
	return "", 0, "Issue with Cookie:" + SessionID
}

func GetUserIDAndFollower2(userName string, SessionID string) (string, int, string) {
	// req, err := http.NewRequest("GET", "https://commentpicker.com/actions/instagram-id-action.php?username="+userName+"&token=29dd11e760c54402744ef2c3273ea2e3c901f88fb4ae749b4882856568ece7b0", nil)
	// if err != nil {
	// 	// handle err
	// }
	// req.Header.Set("Authority", "commentpicker.com")
	// req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.83 Safari/537.36")
	// req.Header.Set("Accept", "*/*")
	// req.Header.Set("Sec-Fetch-Site", "same-origin")
	// req.Header.Set("Sec-Fetch-Mode", "cors")
	// req.Header.Set("Sec-Fetch-Dest", "empty")
	// req.Header.Set("Referer", "https://commentpicker.com/instagram-user-id.php")
	// req.Header.Set("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
	// // req.Header.Set("Cookie", "ezoadgid_186623=-1; ezoref_186623=google.com; ezoab_186623=mod1; ezopvc_186623=1; ezepvv=333; lp_186623=https://commentpicker.com/instagram-user-id.php; ezovid_186623=1668751742; ezovuuid_186623=f0670d60-846a-4a4f-58c9-761391c8c345; ezCMPCCS=true; _ga=GA1.2.1779998933.1599997405; _gid=GA1.2.1575635277.1599997405; _gat=1; ezds=ffid%3D1%2Cw%3D1680%2Ch%3D1050; ezohw=w%3D1680%2Ch%3D916; __utma=131166109.1779998933.1599997405.1599997405.1599997405.1; __utmc=131166109; __utmz=131166109.1599997405.1.1.utmcsr=google|utmccn=(organic)|utmcmd=organic|utmctr=(not%20provided); __utmt_e=1; __utmt_f=1; __utmb=131166109.2.10.1599997405; ezosuigeneris=b728907b403c347d808888c8ecb1b4a9; cto_bidid=6B9wUV9GSURmQUg3bjNub3l3ZXJ2d0tGa21DVllReEZWaXduJTJGcTRsTlNhcHY3JTJGYXA1Y0lKRXBjWTA3MGZMTSUyQkRrV3Z1YllKOVFocmpDRXlSa0pCY01GeU52Z0dFT3NWYW9IYW9RRngxT2dPVERXTSUzRA; cto_bundle=reOv2F9YZ0J6OXg3cWIyS280QXhXMm9ldyUyRm5VbTA1VFpucEFmNSUyRndvUnRsa1pJaUhZS3N5UlFMVThMbiUyRjZNamt4aVM3Y2dVWHpkczRTTGRSWHVWUUljR2w4SVNNandXZzlVQkFJb1Z5SzU3SmZidXBUamJxMElBWTFVZkE5MFUlMkZnN0UwN0lsNHdDWDclMkJkMWZScWVkYWtUeXRnJTNEJTNE; ezux_lpl_186623=1599997405287|2abf5736-322c-4f4c-71cf-a1c27141c03b|false; ezouspvh=180; __gads=ID=4e92f1d7d26e48fb:T=1599997405:S=ALNI_MbRSd0jOFWnARwEXRaByNHhmtr3kQ; ezouspvv=496; ezouspva=8; __qca=P0-1035202418-1599997410395; ezux_et_186623=13; ezux_tos_186623=14; ezux_ifep_186623=true; ezoawesome_186623=commentpicker_com-medrectangle-2/2020-09-13/555335 1599997422248; active_template::186623=pub_site.1599997422; ezosuigenerisc=44e8e79940ca45b4801e7fdfd413e2d4")

	// resp, err := http.DefaultClient.Do(req)
	// if err != nil {
	// 	// handle err
	// 	return "", 0
	// }
	// defer resp.Body.Close()
	// body, err := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(body))
	// var instaID InstaID
	// json.Unmarshal(body, &instaID)
	// return instaID.ID, instaID.Followers

	// Url := "https://www.instagram.com/graphql/query/?query_hash=bfa387b2992c3a52dcbe447467b4b771&variables=%7B%22id%22%3A%22" + userName + "%22%2C%22first%22%3A12%7D"
	// fmt.Println(Url)
	// resp, err := soup.Get(Url)
	// if err != nil {
	// 	fmt.Println("username not found")
	// }

	Url := "http://www.instagram.com/" + userName + "/"
	req, err := http.NewRequest("GET", Url, nil)
	if err != nil {
		// handle err
	}
	req.Header.Set("Authority", "www.instagram.com")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 9; SM-A102U Build/PPR1.180610.011; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/74.0.3729.136 Mobile Safari/537.36 Instagram 155.0.0.37.107 Android (28/9; 320dpi; 720x1468; samsung; SM-A102U; a10e; exynos7885; en_US; 239490550)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Add("Cookie", "sessionid="+SessionID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
		return "", 0, "Issue with Cookie:" + SessionID
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	if err != nil {
		fmt.Println("Error")
		return "", 0, "Issue with Cookie:" + SessionID
	} else {
		actual := strings.Index(string(body), "<script type=\"text/javascript\">window._sharedData")
		if actual == -1 {
			return "", 0, "Issue with Cookie:" + SessionID
		}
		end := strings.Index(string(body), "<script type=\"text/javascript\">window.__initialDataLoaded(window._sharedData);</script>")
		if end == -1 {
			return "", 0, "Issue with Cookie:" + SessionID
		}
		filteredString := (string(body)[actual+len("<script type=\"text/javascript\">window._sharedData")+2 : end-11])
		if filteredString == "" {
			return "", 0, "Issue with Cookie:" + SessionID
		}
		var igResponse IGResponse
		json.Unmarshal([]byte(filteredString), &igResponse)
		if len(igResponse.EntryData.ProfilePage) > 0 {
			return igResponse.EntryData.ProfilePage[0].Graphql.User.ID, igResponse.EntryData.ProfilePage[0].Graphql.User.EdgeFollowedBy.Count, ""
		}
	}
	return "", 0, "Issue with Cookie:" + SessionID
}

func GetFollowers(userName string, MaxFollowers string, SessionID string) ([]string, string) {
	MaxFollowersInt, err := strconv.Atoi(MaxFollowers)
	if err != nil {
		MaxFollowersInt = 500
	}
	MaxFollowersCount := 0
	if SessionID == "" {
		SessionID = configs.Configurations.SessionId
	}
	UserID, _, CookieErrorString := GetUserIDAndFollower(userName, GetRandomCookie(SessionID))
	fmt.Println("*****")
	fmt.Println("User Id found: " + UserID)
	fmt.Println("*****")
	if UserID == "" {
		return nil, CookieErrorString
	}
	var finalValues []string
	var igFollowersResearch IGFollowersResearch
	EndCursor := "first"
	NextPage := true
	for NextPage && MaxFollowersCount < MaxFollowersInt {
		var URL string
		if EndCursor == "first" {
			URL = "https://www.instagram.com/graphql/query/?query_hash=d04b0a864b4b54837c0d870b0e77e076&variables=%7B%22id%22%3A%22" + UserID + "%22%2C%22include_reel%22%3Afalse%2C%22fetch_mutual%22%3Afalse%2C%22first%22%3A24%7D"
		} else {
			URL = "https://www.instagram.com/graphql/query/?query_hash=d04b0a864b4b54837c0d870b0e77e076&variables=%7B%22id%22%3A%22" + UserID + "%22%2C%22include_reel%22%3Afalse%2C%22fetch_mutual%22%3Afalse%2C%22first%22%3A12%2C%22after%22%3A%22" + EndCursor[:len(EndCursor)-2] + "%3D%3D%22%7D"
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
		if SessionID == "" {
			SessionID = configs.Configurations.SessionId
		}
		req.Header.Add("accept", " */*")
		req.Header.Add("Cookie", "sessionid="+GetRandomCookie(SessionID))

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
	return finalValues, ""
}

func GetNewFollowers(userName string, LastFetchedFollowers string, SessionID string) ([]string, string, int) {
	LastFetchedFollowers = strings.Replace(LastFetchedFollowers, ",", "", -1)
	LastFetchedFollowersInt, err := strconv.Atoi(LastFetchedFollowers)
	if err != nil {
		LastFetchedFollowersInt = 10
	}
	if SessionID == "" {
		SessionID = configs.Configurations.SessionId
	}
	UserID, _, CookieErrorString := GetUserIDAndFollower(userName, GetRandomCookie(SessionID))
	fmt.Println("*****")
	fmt.Println("User Id found: " + UserID)
	fmt.Println("*****")
	if UserID == "" {
		return nil, CookieErrorString, 0
	}
	var finalValues []string
	var igFollowersResearch IGFollowersResearch
	EndCursor := "first"
	Firstpage := true
	BreakFromAllLoops := false
	NextPage := true
	LatestFollowerCount := 0
	var NumberOfFollowersNeeded int
	for NextPage {
		var URL string
		if EndCursor == "first" {
			URL = "https://www.instagram.com/graphql/query/?query_hash=d04b0a864b4b54837c0d870b0e77e076&variables=%7B%22id%22%3A%22" + UserID + "%22%2C%22include_reel%22%3Afalse%2C%22fetch_mutual%22%3Afalse%2C%22first%22%3A24%7D"
		} else {
			URL = "https://www.instagram.com/graphql/query/?query_hash=d04b0a864b4b54837c0d870b0e77e076&variables=%7B%22id%22%3A%22" + UserID + "%22%2C%22include_reel%22%3Afalse%2C%22fetch_mutual%22%3Afalse%2C%22first%22%3A12%2C%22after%22%3A%22" + EndCursor[:len(EndCursor)-2] + "%3D%3D%22%7D"
		}
		fmt.Println(URL)
		method := "GET"
		client := &http.Client{}
		req, err := http.NewRequest(method, URL, nil)
		if err != nil {
			fmt.Println(err)
			return nil, CookieErrorString, 0
		}
		if SessionID == "" {
			SessionID = configs.Configurations.SessionId
		}
		req.Header.Add("accept", " */*")
		req.Header.Add("Cookie", "sessionid="+GetRandomCookie(SessionID))

		res, err := client.Do(req)
		if err != nil {
			fmt.Println("username not found")
			return nil, CookieErrorString, 0
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("username not found")
			return nil, CookieErrorString, 0
		}
		json.Unmarshal([]byte(body), &igFollowersResearch)
		iterator := 0
		if Firstpage {
			LatestFollowerCount = igFollowersResearch.Data.User.EdgeFollow.Count
			NumberOfFollowersNeeded = LatestFollowerCount - LastFetchedFollowersInt
			if NumberOfFollowersNeeded > 50 {
				NumberOfFollowersNeeded = 10
			}
			fmt.Println(igFollowersResearch)
			fmt.Println("Latest Follower Counts are: ")
			fmt.Println(LatestFollowerCount)
			fmt.Println(LastFetchedFollowersInt)
			fmt.Println("Number Of Top Followers Needed are:")
			fmt.Println(NumberOfFollowersNeeded)
			Firstpage = false
		}
		for iterator < len(igFollowersResearch.Data.User.EdgeFollow.Edges) {
			if igFollowersResearch.Data.User.EdgeFollow.Edges[iterator].Node.Username != "" && NumberOfFollowersNeeded > 0 {
				finalValues = append(finalValues, igFollowersResearch.Data.User.EdgeFollow.Edges[iterator].Node.Username)
				iterator++
				NumberOfFollowersNeeded--
			} else {
				BreakFromAllLoops = true
				break
			}
		}
		EndCursor = igFollowersResearch.Data.User.EdgeFollow.PageInfo.EndCursor
		NextPage = igFollowersResearch.Data.User.EdgeFollow.PageInfo.HasNextPage
		if BreakFromAllLoops {
			break
		}
	}
	return finalValues, "", LatestFollowerCount
}

func GetIGReport(userNames []string, SearchQuery map[string]int) [][]interface{} {
	var finalValues [][]interface{}
	parentIterator := 0
	for parentIterator < len(userNames) {
		var row []interface{}
		time.Sleep(500 * time.Millisecond)
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
		req.Header.Set("Cookie", "ig_did=2E8DBEA9-6BAB-4214-BE14-3E92C1956C79; mid=X2Cs0AAEAAH4q10wWRKpkOR7Vcxk; csrftoken=85768r6cbvT6MHcJ7JXRjAz30M7ZyWWP; ds_user_id=41670979469; sessionid=41670979469%3AXIijRyjzHto0c7%3A26; rur=PRN;")

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
			if actual == -1 {
				continue
			}
			end := strings.Index(string(body), "<script type=\"text/javascript\">window.__initialDataLoaded(window._sharedData);</script>")
			if actual == -1 {
				continue
			}
			filteredString := (string(body)[actual+len("<script type=\"text/javascript\">window._sharedData")+2 : end-11])
			if filteredString == "" {
				continue
			}
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

func GetIGReportNew(userNames []string, SearchQuery map[string]int, SessionID string, NDelta float64) ([][]interface{}, bool, string) {
	var finalValues [][]interface{}
	parentIterator := 0
	NoOneSucceeded := 0
	var CookieErrorString string
	for parentIterator < len(userNames) {
		FirstPage := true
		var row []interface{}
		time.Sleep(1000 * time.Millisecond)
		UserId, _, CookieErrorString := GetUserIDAndFollower(userNames[parentIterator], GetRandomCookie(SessionID))
		Url := "https://www.instagram.com/graphql/query/?query_hash=bfa387b2992c3a52dcbe447467b4b771&variables=%7B%22id%22%3A%22" + UserId + "%22%2C%22first%22%3A12%7D"
		fmt.Println(Url)
		if SessionID == "" {
			SessionID = configs.Configurations.SessionId
		}
		method := "GET"
		client := &http.Client{}
		req, err := http.NewRequest(method, Url, nil)
		if err != nil {
			fmt.Println(err)
		}
		req.Header.Add("accept", " */*")
		req.Header.Add("Cookie", "sessionid="+GetRandomCookie(SessionID))
		res, err := client.Do(req)
		if err != nil {
			fmt.Println("username not found")
		}
		defer res.Body.Close()
		resp, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("username not found")
		}
		var igResponse AutoGenerated
		json.Unmarshal([]byte(resp), &igResponse)
		TotalLikes := 0.0
		TotalComments := 0.0
		StandardDeviation := 0.0
		Variance := 0.0
		NumberOfPostsOnFirstPage := 12
		if FirstPage {
			NumberOfPostsOnFirstPage = len(igResponse.Data.User.EdgeOwnerToTimelineMedia.Edges)
			FirstPage = false
		}
		if len(igResponse.Data.User.EdgeOwnerToTimelineMedia.Edges) > 0 {
			FollowerURL := "https://www.instagram.com/graphql/query/?query_hash=c76146de99bb02f6415203be841dd25a&variables=%7B%22id%22%3A%22" + UserId + "%22%2C%22include_reel%22%3Afalse%2C%22fetch_mutual%22%3Afalse%2C%22first%22%3A24%7D"
			fmt.Println(FollowerURL)
			method := "GET"
			client := &http.Client{}
			req, err := http.NewRequest(method, FollowerURL, nil)
			if err != nil {
				fmt.Println(err)
			}
			req.Header.Add("accept", " */*")
			if SessionID == "" {
				SessionID = configs.Configurations.SessionId
			}
			Cookie := GetRandomCookie(SessionID)
			fmt.Println(CookieErrorString)
			fmt.Println(SessionID)
			req.Header.Add("Cookie", "sessionid="+Cookie)
			//			req.Header.Add("Cookie", "ig_did=2E8DBEA9-6BAB-4214-BE14-3E92C1956C79; mid=X2Cs0AAEAAH4q10wWRKpkOR7Vcxk; csrftoken=85768r6cbvT6MHcJ7JXRjAz30M7ZyWWP; ds_user_id=41670979469; sessionid=41670979469%3AXIijRyjzHto0c7%3A26; rur=PRN;")
			res, err := client.Do(req)
			if err != nil {
				fmt.Println("username not found")
			}
			defer res.Body.Close()
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				fmt.Println("username not found")
			}
			var igFollowersResearch IGFollowersAllResearch
			json.Unmarshal([]byte(body), &igFollowersResearch)

			Followers := igFollowersResearch.Data.User.EdgeFollowedBy.Count
			if val, ok := SearchQuery["MinFollower"]; ok {
				if Followers < val {
					parentIterator++
					NoOneSucceeded++
					continue
				}
			}
			if val, ok := SearchQuery["MaxFollower"]; ok {
				if Followers > val {
					parentIterator++
					NoOneSucceeded++
					continue
				}
			}

			i := 0
			Engagement := make([]int, 12)
			for i < 12 && i < len(igResponse.Data.User.EdgeOwnerToTimelineMedia.Edges) {
				Likes := igResponse.Data.User.EdgeOwnerToTimelineMedia.Edges[i].Node.EdgeMediaPreviewLike.Count
				Comments := igResponse.Data.User.EdgeOwnerToTimelineMedia.Edges[i].Node.EdgeMediaToComment.Count
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
			BestEngagementFloat := float64(total) / (9 * float64(Followers))
			BestEngagement := BestEngagementFloat * 100
			avgEngagementFloat := (float64(TotalLikes) + float64(TotalComments)) / (12 * float64(Followers))
			avgEngagement := avgEngagementFloat * 100

			BestEngagementForSD := (float64(TotalLikes) + float64(TotalComments)) / (float64(NumberOfPostsOnFirstPage) * float64(Followers))

			i = 0
			var sd float64
			for i < 12 {
				sd += math.Pow(((float64(Engagement[i]) / float64(Followers)) - BestEngagementForSD), 2)
				i++
			}
			Variance = sd / (float64(NumberOfPostsOnFirstPage))
			StandardDeviation = math.Sqrt(Variance)
			fmt.Println(float64(Followers))
			fmt.Println(BestEngagement)
			fmt.Println(avgEngagement)
			if val, ok := SearchQuery["MinN"]; ok {
				avgEngagementInt := int(math.Round(avgEngagement))
				if avgEngagementInt < val {
					parentIterator++
					NoOneSucceeded++
					continue
				}
			}
			if val, ok := SearchQuery["MinNStar"]; ok {
				BestEngagementInt := int(math.Round(BestEngagement))
				if BestEngagementInt < val {
					parentIterator++
					NoOneSucceeded++
					continue
				}
			}
			if NDelta > 0.0 {
				if BestEngagement-avgEngagement < NDelta {
					parentIterator++
					NoOneSucceeded++
					continue
				}
			}
			row = append(row, userNames[parentIterator], "https://www.instagram.com/"+userNames[parentIterator], "", Followers, avgEngagementFloat, BestEngagementFloat, BestEngagementFloat-avgEngagementFloat, Variance, StandardDeviation)
		}
		finalValues = append(finalValues, row)
		parentIterator++
	}
	fmt.Println(finalValues)
	var NoOneSucceededBoolean bool
	if len(userNames) == NoOneSucceeded {
		NoOneSucceededBoolean = true
	}

	return finalValues, NoOneSucceededBoolean, CookieErrorString
}

func SessionIDChecker(SessionID string) bool {
	Url := "https://www.instagram.com/web/search/topsearch/?context=blended&query=" + "virat.kohli" + "&rank_token=0.7584840628946048"
	req, err := http.NewRequest("GET", Url, nil)
	if err != nil {
		// handle err
	}
	req.Header.Set("Authority", "www.instagram.com")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 9; SM-A102U Build/PPR1.180610.011; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/74.0.3729.136 Mobile Safari/537.36 Instagram 155.0.0.37.107 Android (28/9; 320dpi; 720x1468; samsung; SM-A102U; a10e; exynos7885; en_US; 239490550)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Add("Cookie", "sessionid="+SessionID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var newInstaID NewInstaID
	json.Unmarshal([]byte(body), &newInstaID)
	if len(newInstaID.Users) > 0 {
		if len(newInstaID.Users[0].User.Pk) > 0 {
			fmt.Println(newInstaID.Users[0].User.Pk)
			if (newInstaID.Users[0].User.Pk) != "" {
				return true
			}
		}
	}
	return false
}

func GetRandomCookie(SessionID string) string {
	IndividualCookie := strings.Split(SessionID, ",")
	Index := 0
	if len(IndividualCookie) > 1 {
		x1 := rand.NewSource(time.Now().UnixNano())
		y1 := rand.New(x1)
		Index = y1.Intn(len(IndividualCookie))
	} else {
		return SessionID
	}
	return IndividualCookie[Index]
}

type AutoGenerated struct {
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
						IsVideo               bool   `json:"is_video"`
						MediaPreview          string `json:"media_preview"`
						TrackingToken         string `json:"tracking_token"`
						EdgeMediaToTaggedUser struct {
							Edges []struct {
								Node struct {
									User struct {
										FullName      string `json:"full_name"`
										ID            string `json:"id"`
										IsVerified    bool   `json:"is_verified"`
										ProfilePicURL string `json:"profile_pic_url"`
										Username      string `json:"username"`
									} `json:"user"`
									X float64 `json:"x"`
									Y float64 `json:"y"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"edge_media_to_tagged_user"`
						DashInfo struct {
							IsDashEligible    bool        `json:"is_dash_eligible"`
							VideoDashManifest interface{} `json:"video_dash_manifest"`
							NumberOfQualities int         `json:"number_of_qualities"`
						} `json:"dash_info"`
						HasAudio           bool   `json:"has_audio"`
						VideoURL           string `json:"video_url"`
						VideoViewCount     int    `json:"video_view_count"`
						EdgeMediaToCaption struct {
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
						ProductType string `json:"product_type"`
					} `json:"node,omitempty"`
				} `json:"edges"`
			} `json:"edge_owner_to_timeline_media"`
		} `json:"user"`
	} `json:"data"`
	Status string `json:"status"`
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

type IGFollowersAllResearch struct {
	Data struct {
		User struct {
			EdgeFollowedBy struct {
				Count int `json:"count"`
			} `json:"edge_followed_by"`
		} `json:"user"`
	} `json:"data"`
}

type IGFollowingAllResearch struct {
	Data struct {
		User struct {
			EdgeFollow struct {
				Count int `json:"count"`
			} `json:"edge_follow"`
		} `json:"user"`
	} `json:"data"`
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
