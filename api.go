package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)
var Db *gorm.DB

type Score struct {
	gorm.Model
	PlayerId uint
	Score uint
}
const connectionString = "REPLACE_WITH_YOUR_DB_URL"

// Set Db, call in init function
func ConnectDatabase() error {
	var err error
	Db, err = gorm.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("error in connectDatabase(): %v", err)
	}
	Db.AutoMigrate(&Score{})
	return err
}

// Push a single score
func PushScore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	decoder := json.NewDecoder(r.Body)
	var score Score
	err := decoder.Decode(&score)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Wrong body schema"))
		return
	}
	Db.Create(&score)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// Clamp top query to not exceed max and only if query present in r
func clampTopQueryParameter(top *int, r *http.Request, max int){
	topQuery := r.URL.Query().Get("top")
	if topQuery != ""{
		*top, _ = strconv.Atoi(topQuery)
		if *top > max {
			*top = max
		}
	}
}

// Get top score of all player
func GetTopScore(w http.ResponseWriter, r *http.Request) {
	var scores []Score
	top := 10
	clampTopQueryParameter(&top, r, 100)

	Db.Order("Score desc").Limit(top).Find(&scores)
	w.Header().Set("Content-Type", "application/json")
	parsedScore, _ := json.Marshal(&scores)
	w.Write([]byte(parsedScore))
}

// Get top score of given player
func GetTopScorePlayer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	playerId, err := strconv.Atoi(params["playerId"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`Not an integer playerId`))
		return
	}
	var scores []Score
	top := 10
	clampTopQueryParameter(&top, r, 1000)

	Db.Order("Score desc").Where(&Score{PlayerId: uint(playerId)}).Limit(top).Find(&scores)
	w.Header().Set("Content-Type", "application/json")
	parsedScore, _ := json.Marshal(&scores)
	w.Write([]byte(parsedScore))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/score", PushScore).Methods(http.MethodPost)
	r.HandleFunc("/score/{playerId:[0-9]+}", GetTopScorePlayer).Methods(http.MethodPost).Queries("top", "{top:[0-9]+}")
	r.HandleFunc("/score/{playerId:[0-9]+}", GetTopScorePlayer)
	r.HandleFunc("/score/best", GetTopScore).Methods(http.MethodGet).Queries("top", "{top:[0-9]+}")
	r.HandleFunc("/score/best", GetTopScore).Methods(http.MethodGet)

	err := ConnectDatabase()
	if err != nil {
		panic("Could not connect to database")
	}

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}