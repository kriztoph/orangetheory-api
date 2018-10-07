package main

import (
	"github.com/gin-gonic/gin"
	"time"
	//"strconv"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/gocolly/colly"
	"encoding/json"
	"os"
	"log"
	"net/http"
	"strconv"
	"net/url"
	"bytes"
	"io/ioutil"
	"math/rand"
)

type class struct {
	Time       string `json:"time"`
	Instructor string `json:"instructor"`
	Waitlist   bool   `json:"waitlisted"`
	Booked bool `json:"booked"`
	InstanceID string `json:"instance_id"`
}

var redisClient *redis.Client

func main() {
	redisHost, ok := os.LookupEnv("REDIS_HOST")
	if !ok {
		log.Fatal("must set REDIS_HOST environment var")
	}

	redisPassword, ok := os.LookupEnv("REDIS_PASSWORD")
	if !ok {
		log.Fatal("must set REDIS_PASSWORD environment var")
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPassword,
		DB:       0,
	})

	r := gin.Default()
	r.GET("/dates", func(c *gin.Context) {
		var dates []string

		date := time.Now()
		dates = append(dates, fmt.Sprintf("%02d/%02d/%02d (%s)", date.Month(), date.Day(), date.Year(), date.Weekday().String()[:3]))

		date = date.AddDate(0, 0, 1)
		dates = append(dates, fmt.Sprintf("%02d/%02d/%02d (%s)", date.Month(), date.Day(), date.Year(), date.Weekday().String()[:3]))

		date = date.AddDate(0, 0, 1)
		dates = append(dates, fmt.Sprintf("%02d/%02d/%02d (%s)", date.Month(), date.Day(), date.Year(), date.Weekday().String()[:3]))

		date = date.AddDate(0, 0, 1)
		dates = append(dates, fmt.Sprintf("%02d/%02d/%02d (%s)", date.Month(), date.Day(), date.Year(), date.Weekday().String()[:3]))

		date = date.AddDate(0, 0, 1)
		dates = append(dates, fmt.Sprintf("%02d/%02d/%02d (%s)", date.Month(), date.Day(), date.Year(), date.Weekday().String()[:3]))

		date = date.AddDate(0, 0, 1)
		dates = append(dates, fmt.Sprintf("%02d/%02d/%02d (%s)", date.Month(), date.Day(), date.Year(), date.Weekday().String()[:3]))

		date = date.AddDate(0, 0, 1)
		dates = append(dates, fmt.Sprintf("%02d/%02d/%02d (%s)", date.Month(), date.Day(), date.Year(), date.Weekday().String()[:3]))

		date = date.AddDate(0, 0, 1)
		dates = append(dates, fmt.Sprintf("%02d/%02d/%02d (%s)", date.Month(), date.Day(), date.Year(), date.Weekday().String()[:3]))

		c.JSON(200, gin.H{
			"dates": dates,
		})
	})

	r.POST("/schedule/:site/refresh", func(c *gin.Context) {
		//site := c.Param("site")
		//now := time.Now()
		//cacheKey := fmt.Sprintf("%s_%d_%d", site, now.Month(), now.Day())

		//requestAndCache(cacheKey, now, site)
	})

	r.POST("/book", func(c *gin.Context) {
		//instanceID := c.Query("instance_id")
		//if instanceID == "" {
		//	c.Status(400)
		//	return
		//}

		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})

			return
		}

		var request map[string]interface{}

		err = json.Unmarshal(body, &request)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})

			return
		}

		var scheduleResponse ScheduleResponse

		err = json.Unmarshal([]byte(request["schedule"].(string)), &scheduleResponse)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})

			return
		}

		classToSchedule := scheduleResponse.ClassMap[request["class"].(string)]
		fmt.Println(classToSchedule)

		authResponse, phpSessionID, err := auth(request["site"].(string), request["username"].(string), request["password"].(string))
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})

			return
		}

		if authResponse.Status != "success" {
			c.Status(401)

			return
		}

		bookClass(classToSchedule.InstanceID, request["site"].(string), phpSessionID)
	})

	r.GET("/schedule/:site", func(c *gin.Context) {
		dateStr := c.Query("date")
		username := c.Query("username")
		if username == "" {
			c.Status(400)
			return
		}

		password := c.Query("password")
		if password == "" {
			c.Status(400)
			return
		}

		date, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})

			return
		}

		site := c.Param("site")

		start := time.Now()
		_, phpSessionID, err := auth(site, username, password)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})

			return
		}
		fmt.Println("time for auth: ", time.Since(start))

		//cacheKey := fmt.Sprintf("%s_%d_%d", site, date.Month(), date.Day())
		//
		//val, err := redisClient.Get(cacheKey).Result()
		//if err == nil {
		//	var times []string
		//
		//	err := json.Unmarshal([]byte(val), &times)
		//	if err != nil {
		//		c.JSON(500, gin.H{
		//			"error": err.Error(),
		//		})
		//
		//		return
		//	}
		//
		//	c.JSON(200, gin.H{"classes": times})
		//
		//	return
		//}

		responseJSON, err := request(date, site, phpSessionID)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(200, responseJSON)
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}

//RandomString - Generate a random string of A-Z chars with len = l
func RandomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25))  //A=65 and Z = 65+25
	}
	return string(bytes)
}

func auth(site string, username string, password string) (*AuthResponse, string, error) {
	params := url.Values{}
	params.Set("email", username)
	params.Set("password", password)
	params.Set("login_ajax", "yes")
	body := bytes.NewBufferString(params.Encode())

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("POST", "https://" + site + ".orangetheoryfitness.com/apps/mindbody/validate_login.php", body)

	// Headers
	phpSessionID := RandomString(15)
	req.Header.Add("Cookie", "__cfduid=dbfe3fa30734cb0130292feccd6a79bd71533687142; _ga=GA1.2.803027176.1533687144; PHPSESSID=" + phpSessionID + "; _gid=GA1.2.449590216.1537746864; ExistingClient=Yes; __ar_v4=6YNUYNX5DVDQFDL774QWIZ%3A20180923%3A7%7CFHBLDFCAJRH7FDGERJNLT2%3A20180923%3A7%7C5FUCRASDRND7NDOV46JLGB%3A20180923%3A7")
	//req.Header.Add("Origin", "https://brentwood-norcal.orangetheoryfitness.com")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9,la;q=0.8")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Accept", "*/*")
	//req.Header.Add("Referer", "https://brentwood-norcal.orangetheoryfitness.com/pages/week-schedule")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Connection", "keep-alive")

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)

	// Display Results
	fmt.Println("response Status : ", resp.Status)
	fmt.Println("response Headers : ", resp.Header)
	fmt.Println("response Body : ", string(respBody))

	var response *AuthResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, "", err
	}

	return response, phpSessionID, nil
}

type AuthResponse struct {
	Status string `json:"status"`
	Name string `json:"name"`
}

type ScheduleResponse struct {
	Classes []class `json:"classes"`
	Response []string `json:"classes_display"`
	ClassMap map[string]class `json:"class_map"`
}

func request(date time.Time, site string, phpSessionID string) (ScheduleResponse, error) {
	//formattedDate := fmt.Sprintf("%02d/%02d/%02d", date.Month(), date.Day(), date.Year())

	classes := []class{}

	cssDate := strconv.Itoa(date.Day())
	c := colly.NewCollector()
	c.OnHTML("#day-class-list-"+cssDate+` li`, func(e *colly.HTMLElement) {
		c := class{
			Time:       e.ChildText(".event-time"),
			Instructor: e.ChildText(".instructor .staff_firstname"),
		}

		waitlistButton := e.ChildText(".waitlist")

		if waitlistButton != "" {
			c.Waitlist = true
		}

		cancelButton := e.ChildText(".cancel_class")
		if cancelButton != "" {
			c.Booked = true
		}

		eventTitle := e.ChildAttrs(".event-title", "id")
		fmt.Println(eventTitle)
		eventTitleID := eventTitle[0]

		instanceAttrs := e.ChildAttrs("#instance_" + eventTitleID, "value")
		fmt.Println(instanceAttrs)

		c.InstanceID = instanceAttrs[0]

		classes = append(classes, c)
	})

	c.OnRequest(func(r *colly.Request) {

	})

	requestDate := fmt.Sprintf("%02d/%02d/%02d", date.Month(), date.Day(), date.Year())
	params := url.Values{}
	params.Set("date", requestDate)
	params.Set("week_view", "0")
	params.Set("show_week", "false")
	params.Set("week_date", "")
	encodedParams := params.Encode()

	start := time.Now()

	c.Request("POST", "https://" + site + ".orangetheoryfitness.com/apps/mindbody/elements/cal.php?" + encodedParams, nil, colly.NewContext(), http.Header{
		"Cookie": []string{"__cfduid=dbfe3fa30734cb0130292feccd6a79bd71533687142; _ga=GA1.2.803027176.1533687144; ExistingClient=Yes; __ar_v4=5FUCRASDRND7NDOV46JLGB%3A20180923%3A26%7CFHBLDFCAJRH7FDGERJNLT2%3A20180923%3A26%7C6YNUYNX5DVDQFDL774QWIZ%3A20180923%3A26; PHPSESSID=" + phpSessionID + "; _gid=GA1.2.1750191254.1538519836; _gat=1"},
		"Accept-Encoding": []string{"gzip, deflate, br"},
		"Accept-Language": []string{"en-US,en;q=0.9,la;q=0.8"},
		"User-Agent": []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36"},
		"Content-Type": []string{"application/x-www-form-urlencoded; charset=UTF-8"},
		"Accept": []string{"*/*"},
		"X-Requested-With": []string{"XMLHttpRequest"},
		"Connection": []string{"keep-alive"},
	})

	fmt.Println(time.Since(start).String())
	c.Wait()

	var response []string
	classMap := make(map[string]class)

	for _, class := range classes {
		listing := class.Time + " -- " + class.Instructor
		if class.Waitlist {
			listing = listing + " 🕑"
		}

		if class.Booked {
			listing = listing + " ✅"
		}

		classMap[listing] = class

		response = append(response, listing)
	}

	//responseJSON, err := json.Marshal(response)
	//if err != nil {
	//	return nil, err
	//}

	//err = redisClient.Set(cacheKey, string(responseJSON), 0).Err()
	//if err != nil {
	//	return nil, err
	//}

	return ScheduleResponse{
		Classes: classes,
		Response: response,
		ClassMap: classMap,
	}, nil
}

func bookClass(instanceID string, site string, phpSessionID string) error {
	// book class (POST https://brentwood-norcal.orangetheoryfitness.com/apps/mindbody/signup_ajax.php)

	params := url.Values{}
	params.Set("instance_id", instanceID)
	params.Set("path", "/pages/week-schedule")
	params.Set("day", "")
	params.Set("redirect_params[show_buttons]", "1")
	params.Set("redirect_params[stay_here]", "1")
	params.Set("redirect_params[show_trainers]", "1")
	params.Set("redirect_params[location_fshow]", "1")
	params.Set("redirect_params[sc_fshow]", "1")
	params.Set("redirect_params[cl_fshow]", "1")
	params.Set("redirect_params[fshow]", "0")
	params.Set("redirect_params[trainer_fshow]", "1")
	params.Set("redirect_params[datepick_fshow]", "1")
	params.Set("redirect_params[print_fshow]", "1")
	params.Set("redirect_params[mini_mode]", "0")
	params.Set("redirect_params[today_color]", "")
	params.Set("redirect_params[this_month_color]", "")
	params.Set("redirect_params[cl][]", "7,14")
	params.Set("redirect_params[sc][]", "22")
	params.Set("redirect_params[locationid]", "all")
	params.Set("redirect_params[date]", "09/23/2018")
	params.Set("redirect_params[week_view]", "1")
	params.Set("redirect_params[show_week]", "0")
	params.Set("redirect_params[month]", "9")
	params.Set("redirect_params[year]", "2018")
	params.Set("day_number", "")
	params.Set("waitlist", "true")
	body := bytes.NewBufferString(params.Encode())

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("POST", "https://" + site + ".orangetheoryfitness.com/apps/mindbody/signup_ajax.php", body)
	if err != nil {
		return err
	}
	// Headers
	req.Header.Add("Cookie", "PHPSESSID=" + phpSessionID)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)

	// Display Results
	fmt.Println("response Status : ", resp.Status)
	fmt.Println("response Headers : ", resp.Header)
	fmt.Println("response Body : ", string(respBody))

	return nil
}