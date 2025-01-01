package main

import (
	"database/sql"
	_ "encoding/json"
	"html/template"
	"log"
	"net/http"
	_ "github.com/mattn/go-sqlite3"
)

type Note struct {
	ID int
	Title string
	Content string
	CreatedAt string
}

var db *sql.DB
func init_db() *sql.DB {
	db, err := sql.Open("sqlite3", "noted.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	create_table_query := `
	CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(create_table_query)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	return db
}

func main() {
	db = init_db()
	defer db.Close()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	
	http.HandleFunc("/", list_notes)
	http.HandleFunc("/note/create", create_note)
	http.HandleFunc("/note/edit/", edit_note)
	http.HandleFunc("/note/delete/", delete_note)

	log.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func list_notes(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, content, created_at FROM notes")
	if err != nil {
		http.Error(w, "Failed to retrieve notes!", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	notes := []Note{}
	for rows.Next() {
		var note Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt); err != nil {
			http.Error(w, "Failed to parse notes!", http.StatusInternalServerError)
			return
		}
		notes = append(notes, note)
	}
	
	tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/index.html"))
	err = tmpl.Execute(w, notes)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Failed to render template!", http.StatusInternalServerError)
	}
}

func create_note(w http.ResponseWriter, r *http.Request) {
	// handle GET request: show the form
	if r.Method == http.MethodGet {
		tmpl := template.Must(template.ParseFiles("templates/base.html","templates/create.html"))
		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, "Failed to render template!", http.StatusInternalServerError)
		}
		return
	}

	// handle POST request: process form
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form data!", http.StatusBadRequest)
			return
		}

		// retrieve form values
		title := r.FormValue("title")
		content := r.FormValue("content")

		// insert into database
		_, err := db.Exec("INSERT INTO NOTES (title, content) VALUES (?, ?)", title, content)
		if err != nil {
			http.Error(w, "Failed to create note!", http.StatusInternalServerError)
			return
		}

		// redirect to homepage
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func edit_note(w http.ResponseWriter, r *http.Request) {
	// extract not ID from URL
	id := r.URL.Path[len("/note/edit/"):]

	// handle GET request: render edit note form
	if r.Method == http.MethodGet {
		var note Note
		err := db.QueryRow("SELECT id, title, content, created_at FROM notes WHERE id = ?", id).Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				http.NotFound(w, r)
			} else {
				http.Error(w, "Failed to retrieve note!", http.StatusInternalServerError)
			}
			return
		}

		tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/edit.html"))
		if err := tmpl.Execute(w, note); err != nil {
			http.Error(w, "Failed to render template!", http.StatusInternalServerError)
		}
		return
	}

	// handle POST request: update the note
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form data!", http.StatusBadRequest)
			return
		}

		// get updated values
		title := r.FormValue("title")
		content := r.FormValue("content")

		// update database
		_, err := db.Exec("UPDATE notes SET title = ?, content = ? WHERE id = ?", title, content, id)
		if err != nil {
			http.Error(w, "Failed to update notes database!", http.StatusInternalServerError)
			return
		}
		
		// redirect to homepage
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func delete_note(w http.ResponseWriter, r *http.Request) {
	// extract ID from URL
	id := r.URL.Path[len("/note/delete/"):]

	// execute delete query
	_, err := db.Exec("DELETE FROM notes WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Failed to delete note!", http.StatusInternalServerError)
		return
	}

	// redirect to homepage
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
