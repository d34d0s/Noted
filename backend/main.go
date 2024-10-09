package main

import (
	"fmt"
	"log"
	"strconv"
	"net/http"
	"database/sql"
	"html/template"

	_ "github.com/ncruces/go-sqlite3/embed"
	_ "github.com/ncruces/go-sqlite3/driver"
)

type Note_t struct {
	ID int
	Name string
	Body string
	Created string
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./backend/db/noted.db")
	if err != nil {
		return nil, err
	}

	q := `
		CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			body TEXT NOT NULL,
			created DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err = db.Exec(q)
	if err != nil {
		return nil, err
	}; return db, nil
}

func queryNotes(db *sql.DB) ([]Note_t, error) {
	rows, err := db.Query("SELECT id, name, body, created FROM notes")
	if err != nil {
		return nil, err
	}; defer rows.Close()

	var notes []Note_t
	for rows.Next() {
		var note Note_t
		if err := rows.Scan(&note.ID, &note.Name, &note.Body, &note.Created); err != nil {
			return nil, err
		}; notes = append(notes, note)
	}; return notes, nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl := template.Must(template.ParseFiles("backend/templates/index.html"))
		tmpl.Execute(w, nil)
	}
}

func notesHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == http.MethodGet || r.Method == http.MethodPost {
		tmpl := template.Must(template.ParseFiles("backend/templates/notes.html"))

		notes, err := queryNotes(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}; tmpl.Execute(w, notes)
	}
}

func addNoteHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == http.MethodPost {
		name := r.FormValue("note-name")
		body := r.FormValue("note-body")

		if name == "" || body == "" {
			http.Error(w, "Note Cannot Be Empty", http.StatusBadRequest)
			return
		}

		_, err := db.Exec("INSERT INTO notes (name, body) VALUES (?, ?)", name, body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}; notesHandler(w, r, db)
	}
}

func delNoteHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == http.MethodPost {
		idStr := r.FormValue("note-id")
		
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid Note ID", http.StatusBadRequest)
			return
		}

		_, err2 := db.Exec("DELETE FROM notes WHERE id = ?", id)
		if err2 != nil {
			http.Error(w, err2.Error(), http.StatusInternalServerError)
			return
		}; notesHandler(w, r, db)
	}
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}; defer db.Close()
	
	staticFiles := http.FileServer(http.Dir("frontend/static"))
	http.Handle("/static/", http.StripPrefix("/static/", staticFiles))

	http.HandleFunc("/", indexHandler)

	http.HandleFunc("/notes", func(w http.ResponseWriter, r *http.Request) {
		notesHandler(w, r, db)
	})
	
	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		addNoteHandler(w, r, db)
	})

	http.HandleFunc("/del", func(w http.ResponseWriter, r *http.Request) {
		delNoteHandler(w, r, db)
	})
		
	fmt.Println("Server Started At :42069")
	http.ListenAndServe(":42069", nil)
}
