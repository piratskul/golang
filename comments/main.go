package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type rawComment struct {
	ID     int
	Text   string
	par_id int
}

type Comment struct {
	ID      int
	Text    string
	Replies []Comment
}

func main() {
	// Get all comments.
	comments := getComments()

	// Delete comment by ID.
	deleteComment(3, &comments)

	// Print out all results.
	for _, com := range comments {
		fmt.Printf("%v, %v, %v\n", com.ID, com.Text, com.Replies)
	}
}

func getComments() []Comment {
	// Connect to database.
	db, err := sql.Open("postgres", "postgres://postgres:1510@localhost/comments?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		panic(err)
	}

	// Get all comments and swap NULL with 0 for parent comments.
	rows, err := db.Query("SELECT id, text, COALESCE(parent_comment_id, '0') FROM comments")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	// Get raw comments.
	rawComs := make([]rawComment, 0)
	for rows.Next() {
		cm := rawComment{}
		err := rows.Scan(&cm.ID, &cm.Text, &cm.par_id)
		if err != nil {
			panic(err)
		}
		rawComs = append(rawComs, cm)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}

	// Create slice of comments with desired structure.
	coms := make([]Comment, 0)
	for _, v := range rawComs {
		if v.par_id == 0 {
			cm := Comment{}
			cm.ID, cm.Text = v.ID, v.Text
			if hasChildren(v.ID, rawComs) {
				cm.Replies = append(cm.Replies, addChildren(v.ID, rawComs)...)
			}
			coms = append(coms, cm)
		}
	}

	return coms
}

// Check if rawComment has children/replies.
func hasChildren(id int, rawComs []rawComment) bool {
	for _, rc := range rawComs {
		if rc.par_id == id {
			return true
		}
	}
	return false
}

// Add children/replies to comment.
func addChildren(id int, rawComs []rawComment) []Comment {
	coms := make([]Comment, 0)
	for _, rc := range rawComs {
		if rc.par_id == id {
			cm := Comment{}
			cm.ID, cm.Text = rc.ID, rc.Text
			if hasChildren(rc.ID, rawComs) {
				cm.Replies = append(cm.Replies, addChildren(rc.ID, rawComs)...)
			}
			coms = append(coms, cm)
		}
	}
	return coms
}

// Remove comment and replies from slice.
func deleteComment(id int, comments *[]Comment) {
	comm := *comments
	for i, c := range comm {
		if c.ID == id {
			*comments = append(comm[:i], comm[i+1:]...)
		}
		if len(c.Replies) > 0 {
			deleteComment(id, &comm[i].Replies)
		}
	}
}
