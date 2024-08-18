package espn

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	client *http.Client
}

type Season struct {
	Events []Event `json:"events"`
}

type Event struct {
	Date         string        `json:"date"`
	ShortName    string        `json:"shortName"`
	Week         Week          `json:"week"`
	Competitions []Competition `json:"competitions"`
	Type         struct {
		Year int    `json:"year"`
		Type int    `json:"type"`
		Slug string `json:"slug"`
	} `json:"season"`
}

type Week struct {
	Number int `json:"number"`
}

type Competition struct {
	Competitors []Competitor `json:"competitors"`
	Odds        []Odd        `json:"odds"`
	ID          string       `json:"id"`
}

type Competitor struct {
	HomeAway string `json:"homeAway"`
	Team     Team   `json:"team"`
}

type Team struct {
	ID       string `json:"id"`
	Location string `json:"location"`
	Name     string `json:"name"`
	Abr      string `json:"abbreviation"`
	Color    string `json:"color"`
	AltColor string `json:"alternateColor"`
	Logo     string `json:"logo"`
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
	ID   string `json:"id"`
	Date string `json:"date"`
	Slug string `json:"slug"`
	Year int    `json:"year"`
	Week int    `json:"week"`
	Type int    `json:"type"`
	Home Team   `json:"home"`
	Away Team   `json:"away"`
	Odds Odd    `json:"odds"`
}

const (
	BASE_API      = "https://site.api.espn.com/apis/site/v2/sports/football/nfl"
	SCHEDULE_PATH = "/scoreboard?limit=1000&dates=20240901-20250228"
	ODDS_API      = "https://sports.core.api.espn.com/v2/sports/football/leagues/nfl/events/%[1]s/competitions/%[1]s/odds"
)

func NewClient() *Client {
	client := http.Client{
		Timeout: time.Second * 10,
	}
	return &Client{client: &client}
}

func (c Client) GetSchedule() (*Schedule, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s%s", BASE_API, SCHEDULE_PATH))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	var season Season
	if err := json.Unmarshal(body, &season); err != nil {
		return nil, err
	}

	games := []Game{}
	for _, e := range season.Events {
		games = append(games, upsertGame(e))
	}
	schedule := &Schedule{Games: games}
	return schedule, nil
}

func (c Client) GetGameOdds(id string) (*Odd, error) {
	resp, err := c.client.Get(fmt.Sprintf(ODDS_API, id))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	type oddsRes struct {
		Items []Odd `json:"items"`
	}
	var odds oddsRes
	if err := json.Unmarshal(body, &odds); err != nil {
		return nil, err
	}

	if len(odds.Items) == 0 {
		return nil, errors.New("no odds found")
	}

	return &odds.Items[0], nil
}

func upsertGame(e Event) Game {
	res := &Game{
		Date: e.Date,
		Year: e.Type.Year,
		Week: e.Week.Number,
		Type: e.Type.Type,
		Slug: e.Type.Slug,
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
		res.ID = cm.ID
	}
	return *res
}
