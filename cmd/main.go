package main

import (
	"net/http"
	"io/ioutil"
	"os"
	"bytes"
	"encoding/json"
	"time"
	"fmt"
	"net/smtp"

	cron "github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

var (
	token string = "lip_WDxlYMj0guVPYDnKI6ox"
	url string = "https://lichess.org/api"
)

type Tournament struct {
	Name 			 string  `json:"name"`
	ClockTime 		 float32  `json:"clockTime"`
	ClockIncrement 	 int 	 `json:"clockIncrement"`
	Minutes 		 int 	 `json:"minutes"`
	WaitMinutes 	 int 	 `json:"waitMinutes"`
	StartDate        int64   `json:"startDate"`
	Variant          string  `json:"variant"`
	Position         string  `json:"position"`
	Rated            bool    `json:"rated"`
	Berserkable      bool    `json:"berserkable"`
	Streakable       bool    `json:"streakable"`
	HasChat          bool    `json:"hasChat"`
	Description      string  `json:"description"`
	Password 		 string  `json:"password"`
	TeamBattleByTeam string  `json:"teamBattleByTeam"`
	TeamID           string  `json:"conditions.teamMember.teamId"`
}

func main() {
	client := &http.Client{}
	c := cron.New()

	host := "smtp.gmail.com"
    port := "587"
	from := "jdap.dmc@gmail.com"
	password := "dtiskgqtfcuavncy"

	toList := []string{"jdap.dmc@gmail.com"}

	file, err := os.OpenFile("logs.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(file)

	c.AddFunc("* 10 * * 5", func () {
		log.Info("Init Cron")

		var data map[string]interface{}

		now := time.Now()

		loc, _ := time.LoadLocation("America/Caracas")

		fixedTime := time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, loc)

		startDate := fixedTime.UnixNano() / 1000000

		payload := Tournament{
			"Torneo de los viernes DCyT",
			5.0,
			3,
			45,
			10,
			startDate,
			"standard",
			"",
			false,
			true,
			true,
			true,
			"",
			"",
			"",
			"",
		}

		reqBody, err := json.Marshal(payload)

        if err != nil {
			log.Panic(err)
        }

		req, err := http.NewRequest(http.MethodPost, url + "/tournament", bytes.NewBuffer(reqBody))

		if err != nil {
			log.Panic(err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer " + token)

		res, err := client.Do(req)

		if err != nil {
			log.Panic(err)
		}

		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)

    	if err != nil {
    		log.Panic(err)
    	}

		e := json.Unmarshal(body, &data)

		if e != nil {
			log.Panic(e)
		}

		id := fmt.Sprintf("%v", data["id"])

		msg := []byte("Subject: Torneo de los viernes \r\nTorneo Blitz de los viernes (5 + 3)\n\nLos Esperamos!!!!\n\nhttps://lichess.org/tournament/" + id)

		auth := smtp.PlainAuth("", from, password, host)

		if err := smtp.SendMail(host + ":" + port, auth, from, toList, msg) ; err != nil {
			log.Panic(err)
		}

		log.Info("Finish Cron")
	})

	c.Start()

	select {}
}
