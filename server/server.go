package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
)

const apiVersion = "1"

type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Battlesnake struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Health         int            `json:"health"`
	Body           []Coord        `json:"body"`
	Head           Coord          `json:"head"`
	Length         int            `json:"length"`
	Latency        string         `json:"latency"`
	Shout          string         `json:"shout"`
	Squad          string         `json:"squad"`
	Customizations Customizations `json:"customizations"`
}

type Customizations struct {
	Color string `json:"color"`
	Head  string `json:"head"`
	Tail  string `json:"tail"`
}

type Board struct {
	Height  int           `json:"height"`
	Width   int           `json:"width"`
	Food    []Coord       `json:"food"`
	Hazards []Coord       `json:"hazards"`
	Snakes  []Battlesnake `json:"snakes"`
}

type GameState struct {
	Game  Game        `json:"game"`
	Turn  int         `json:"turn"`
	Board Board       `json:"board"`
	You   Battlesnake `json:"you"`
}

type Game struct {
	ID      string  `json:"id"`
	Ruleset Ruleset `json:"ruleset"`
	Map     string  `json:"map"`
	Source  string  `json:"source"`
	Timeout int     `json:"timeout"`
}

type Ruleset struct {
	Name     string          `json:"name"`
	Version  string          `json:"version"`
	Settings RulesetSettings `json:"settings"`
}

type RulesetSettings struct {
	FoodSpawnChance     int           `json:"foodSpawnChance"`
	MinimumFood         int           `json:"minimumFood"`
	HazardDamagePerTurn int           `json:"hazardDamagePerTurn"`
	HazardMap           string        `json:"hazardMap"`
	HazardMapAuthor     string        `json:"hazardMapAuthor"`
	Royale              RoyalSettings `json:"royale"`
	Squad               SquadSettings `json:"squad"`
}

type RoyalSettings struct {
	ShrinkEveryNTurns int `json:"shrinkEveryNTurns"`
}

type SquadSettings struct {
	AllowBodyCollisions bool `json:"allowBodyCollisions"`
	SharedElimination   bool `json:"sharedElimination"`
	SharedHealth        bool `json:"sharedHealth"`
	SharedLength        bool `json:"sharedLength"`
}

type InfoResponse struct {
	APIVersion string `json:"apiversion"`
	Author     string `json:"author"`
	Color      string `json:"color"`
	Head       string `json:"head"`
	Tail       string `json:"tail"`
	Version    string `json:"version"`
}

type MoveResponse struct {
	Move  string `json:"move"`
	Shout string `json:"shout"`
}

type battlesnakeServer struct {
	http.Handler
	port     string
	info     *InfoResponse
	logger   *Logger
	moveFunc func(*GameState, *Logger) MoveResponse
}

func New(port string, info *InfoResponse, logger *log.Logger, loggerOpts int, moveFunc func(*GameState, *Logger) MoveResponse) *battlesnakeServer {
	info.APIVersion = apiVersion
	s := &battlesnakeServer{
		port:     port,
		info:     info,
		logger:   NewLogger(logger, loggerOpts),
		moveFunc: moveFunc,
	}
	r := http.NewServeMux()
	r.HandleFunc("/", s.withRequestLogging(s.indexHandler()))
	r.HandleFunc("/start", s.withRequestLogging(s.startHandler()))
	r.HandleFunc("/end", s.withRequestLogging(s.endHandler()))
	r.HandleFunc("/move", s.withRequestLogging(s.moveHandler()))
	s.Handler = r
	return s
}

func (s *battlesnakeServer) Start() error {
	ln, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}
	s.logger.Printf("START server running at %s...\n", ln.Addr())
	s.logger.Debug("request debug logging enabled")
	return http.Serve(ln, s)
}

func (s *battlesnakeServer) indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("content-type", "application/json")
		if err := json.NewEncoder(w).Encode(s.info); err != nil {
			s.logger.Err("Failed to encode index response: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (s *battlesnakeServer) startHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := &GameState{}
		if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
			s.logger.Err("Failed to decode start request body: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		s.logger.Info("Game ID %s [Turn %d] Snake ID %s - Start", state.Game.ID, state.Turn, state.You.ID)
	}
}

func (s *battlesnakeServer) endHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := &GameState{}
		if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
			s.logger.Err("Failed to decode end request body: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		s.logger.Info("Game ID %s [Turn %d] Snake ID %s - End", state.Game.ID, state.Turn, state.You.ID)
	}
}

func (s *battlesnakeServer) moveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := &GameState{}
		if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
			s.logger.Err("Failed to decode move request body: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp := s.moveFunc(state, s.logger)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Err("Failed to encode move response, %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		s.logger.Info("Game ID %s [Turn %d] Snake ID %s - Move: %s", state.Game.ID, state.Turn, state.You.ID, resp.Move)
	}
}

func (s *battlesnakeServer) withRequestLogging(next http.HandlerFunc) http.HandlerFunc {
	if s.logger.Enabled(LDebug) {
		return func(w http.ResponseWriter, r *http.Request) {
			b, err := io.ReadAll(r.Body)
			if err != nil {
				s.logger.Err("Failed to read request body: %s", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(b))
			s.logger.Debug("%s %s %s\n", r.Method, r.URL.Path, string(b))
			next(w, r)
		}
	}
	return next
}
