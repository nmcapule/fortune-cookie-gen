package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Resp struct {
	Error  int         `json:"id"`
	Output interface{} `json:"output"`
}

type Cookie struct {
	ID           int       `json:"id"`
	Message      string    `json:"message"`
	Passes       int       `json:"passes"`
	CreatedDate  time.Time `db:"created_date" json:"created_date"`
	ModifiedDate time.Time `db:"modified_date" json:"last_modified"`
}

var (
	templates *template.Template
	db        *sqlx.DB
)

var (
	DropTable   = `DROP TABLE %v`
	TableCookie = `
CREATE TABLE cookie (
  id SERIAL,
  message varchar(160) NOT NULL DEFAULT '',
  passes numeric(11) NOT NULL DEFAULT 0,
  created_date timestamp,
  modified_date timestamp,
  PRIMARY KEY (id)
)
`
	RowSelectCookie = `
SELECT
  id, message, passes, created_date, modified_date
FROM cookie
`
	RowInsertCookie = `
INSERT INTO cookie (
  message, passes, created_date, modified_date
) VALUES ($1, $2, $3, $4)
`
	RowUpdateCookieCounter = `
UPDATE cookie SET passes = $1 WHERE id = $2
`
	RowDeleteCookie = `
DELETE FROM cookie
WHERE id = $1
`
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getCookiesHandler(w http.ResponseWriter, r *http.Request) {
	var cookies []Cookie

	err := db.Select(&cookies, RowSelectCookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(cookies)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getCookieHandler(w http.ResponseWriter, r *http.Request) {
	var cookie Cookie

	query := RowSelectCookie + " WHERE id = $1"
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	err := db.Get(&cookie, query, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(RowUpdateCookieCounter, cookie.Passes+1, cookie.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(cookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func putCookieHandler(w http.ResponseWriter, r *http.Request) {
	var cookie Cookie

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	_, err := db.Exec(RowDeleteCookie, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&cookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	cookie.ModifiedDate = time.Now()
	cookie.CreatedDate = time.Now()
	_, err = db.Exec(RowInsertCookie, cookie.Message, cookie.Passes, time.Now(), time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(cookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func postCookieHandler(w http.ResponseWriter, r *http.Request) {
	var cookie Cookie

	err := json.NewDecoder(r.Body).Decode(&cookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	cookie.ModifiedDate = time.Now()
	cookie.CreatedDate = time.Now()
	_, err = db.Exec(RowInsertCookie, cookie.Message, cookie.Passes, time.Now(), time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(cookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func setupSchema() error {
	tables := map[string]string{
		"cookie": TableCookie,
	}

	tx, _ := db.Beginx()
	defer tx.Commit()

	for table, q := range tables {
		_, err := db.Exec(fmt.Sprintf(DropTable, table))
		if err != nil {
			return err
		}

		_, err = db.Exec(q)
		if err != nil {
			return err
		}
	}

	return nil

}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	err := setupSchema()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Success")
}

func main() {
	var err error

	db, err = sqlx.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	templates = template.Must(template.ParseFiles("templates/index.html"))

	r := mux.NewRouter()

	r.HandleFunc("/", mainHandler)
	r.HandleFunc("/sap", getCookiesHandler)
	r.HandleFunc("/sap/{id:[0-9]+}", getCookieHandler)
	r.HandleFunc("/up", postCookieHandler)
	r.HandleFunc("/matoyo", adminHandler)
	http.Handle("/", r)

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}
	log.Printf("Serving app at %v", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))

}
