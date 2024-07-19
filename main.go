package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Season struct {
	Events []Event `json:"events"`
}

type Event struct {
	Date         string        `json:"date"`
	ShortName    string        `json:"shortName"`
	Week         Week          `json:"week"`
	Competitions []Competition `json:"competitions"`
}

type Week struct {
	Number int `json:"number"`
}

type Competition struct {
	Competitors []Competitor `json:"competitors"`
	Odds        []Odd        `json:"odds"`
}

type Competitor struct {
	HomeAway string `json:"homeAway"`
	Team     Team   `json:"team"`
}

type Team struct {
	Location string `json:"location"`
	Name     string `json:"name"`
	Abr      string `json:"abbreviation"`
	Color    string `json:"color"`
	AltColor string `json:"alternateColor"`
}

type Odd struct {
	Details   string  `json:"details"`
	OverUnder float32 `json:"overUnder"`
	Spread    float32 `json:"spread"`
}

type Schedule struct {
	Games []Game `json:"games"`
}

type Game struct {
	Date string `json:"date"`
	Week int    `json:"week"`
	Home Team   `json:"home"`
	Away Team   `json:"away"`
	Odds Odd    `json:"odds"`
}

func main() {

	const (
		API = "https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard"
	)

	resp, err := http.Get(API)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body) // response body is []byte

	var season Season
	if err := json.Unmarshal(body, &season); err != nil { // Parse []byte to the go struct pointer
		fmt.Println(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		games := []Game{}
		for _, e := range season.Events {
			game := getGame(e)
			games = append(games, game)
		}
		res := Schedule{Games: games}
		json.NewEncoder(w).Encode(res)
	})

	fmt.Println("Starting server at port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server failed:", err)
	}
}

func getGame(e Event) Game {
	res := &Game{
		Date: e.Date,
		Week: e.Week.Number,
	}
	for _, cm := range e.Competitions {
		for _, c := range cm.Competitors {
			if c.HomeAway == "home" {
				res.Home = c.Team
			} else {
				res.Away = c.Team
			}
		}
		for _, o := range cm.Odds {
			res.Odds = o
		}
	}
	return *res
}
