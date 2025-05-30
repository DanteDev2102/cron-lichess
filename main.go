package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"sync"
	"time"

	godotenv "github.com/joho/godotenv"
	cron "github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

var mutex sync.Mutex
var tournaments map[string]bool

type Config struct {
	LichessToken               string
	LichessURL                 string
	CronPeriod                 string
	SMTPHost                   string
	SMTPPort                   string
	SMTPFrom                   string
	SMTPPassword               string
	SMTPTarget                 string
	TournamentName             string
	TournamentTimezone         string
	TournamentMinutes          int
	TournamentVariants         []string
	TournamentDescription      string
	TournamentHasChat     	   bool 
	TournamentBerserkable      bool
	TournamentRated            bool
	TournamentStreakable       bool
	TournamentWaitMinutes      int
	TournamentPassword         string
	TournamentTeamBattleByTeam string
	TournamentTeamID           string
}

type Tournament struct {
	Name             string  `json:"name"`
	ClockTime        float32 `json:"clockTime"`
	ClockIncrement   int     `json:"clockIncrement"`
	Minutes          int     `json:"minutes"`
	WaitMinutes      int     `json:"waitMinutes"`
	StartDate        int64   `json:"startDate"`
	Variant          string  `json:"variant"`
	Position         string  `json:"position"`
	Rated            bool    `json:"rated"`
	Berserkable      bool    `json:"berserkable"`
	Streakable       bool    `json:"streakable"`
	HasChat          bool    `json:"hasChat"`
	Description      string  `json:"description"`
	Password         string  `json:"password"`
	TeamBattleByTeam string  `json:"teamBattleByTeam"`
	TeamID           string  `json:"conditions.teamMember.teamId"`
}

func createTournament(startDate int64) Tournament {
	times := []float32{1.0, 3.0, 5.0, 8.0, 10.0}
	increments := []int{0, 1, 2, 3, 5}
	variants := []string{"standard", "chess960", "crazyhouse", "horde", "racingKings"}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randIndex := r.Intn(len(times))

	return Tournament{
		Name:             "Torneo de los Viernes DCyT",
		ClockTime:        times[randIndex],
		ClockIncrement:   increments[randIndex],
		Minutes:          60,
		WaitMinutes:      10,
		StartDate:        startDate,
		Variant:          variants[randIndex],
		Position:         "",
		Rated:            false,
		Berserkable:      true,
		Streakable:       true,
		HasChat:          true,
		Description:      "",
		Password:         "",
		TeamBattleByTeam: "",
		TeamID:           "",
	}
}

func sendTournament(tournament Tournament) ([]byte, error) {
	url := os.Getenv("LICHESS_URL")
	token := os.Getenv("LICHESS_TOKEN")
	client := &http.Client{}

	body, err := json.Marshal(tournament)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url+"/tournament", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	return data, nil

}

func sendMail(id string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASS")
	target := os.Getenv("SMTP_TARGET")

	toList := []string{target}

	msg := []byte("https://lichess.org/tournament/" + id)

	auth := smtp.PlainAuth("", from, password, host)

	return smtp.SendMail(host+":"+port, auth, from, toList, msg)

}



func (c *Config) load() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	
	
}

func generateStartDate(now time.Time) int64 {
	loc, _ := time.LoadLocation("America/Caracas")
	fixedTime := time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, loc)
	return fixedTime.UnixNano() / 1000000
}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Panic(err)
	}

	c := cron.New()

	period := os.Getenv("CRON_PERIOD")

	file, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(file)

	val, err := c.AddFunc(period, func() {
		log.Info("Init Cron")
		fmt.Println("Init Cron")

		var data map[string]interface{}

		now := time.Now()
		startDate := generateStartDate(now)
		
		fmt.Println("Start Date:", startDate)

		mutex.Lock()
		key := fmt.Sprintf("%d/%d/%d", now.Year(), now.Month(), now.Day())
		_, ok := tournaments[key]

		if ok {
			mutex.Unlock()
			return
		}

		tournaments = make(map[string]bool)
		tournaments[key] = true
		mutex.Unlock()

		payload := createTournament(startDate)
		body, err := sendTournament(payload)
		if err != nil {
			log.Error(err)
			return
		}

		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Error(err)
			return
		}

		id := fmt.Sprintf("%v", data["id"])
		err = sendMail(id)
		if err != nil {
			log.Error(err)
			return
		}

		log.Info("Finish Cron")
	})

	if err != nil {
		log.Error("Error adding cron job:", err)
		return
	}
	log.Info("Cron job added with ID:", val)

	c.Start()

	select {}
}
