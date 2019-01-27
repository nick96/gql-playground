package common

import (
	"database/sql"
	"log"
	"fmt"

	// Import the sqlite3 library so we abstract away from it
	_ "github.com/mattn/go-sqlite3"
)

// TodoItem is a representation of an item in a todo list.
type TodoItem struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Done   bool   `json:"done"`
	ListID int    `json:"listId"`
}

// TodoList ...
type TodoList struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// DB is the database connection
var db *sql.DB

func init() {
	db, err := sql.Open("sqlite3", "todo.db")
	if err != nil {
		log.Fatalf("could not open database, error: %v", err)
	}

	todoListsTable := `
	CREATE TABLE IF NOT EXISTS todoLists (
		id INTEGER PRIMARY KEY autoincrement,
		name TEXT
	);`

	todoItemsTable := `
	CREATE TABLE IF NOT EXISTS todoItems (
		id INTEGER PRIMARY KEY autoincrement,
		name TEXT,
		done BOOLEAN,
		listId INTEGER,
		FOREIGN KEY (listId) REFERENCES todoLists(id)
	);`

	_, err = db.Exec(todoListsTable)
	if err != nil {
		log.Fatalf("could not create todo lists table, error: %v", err)
	}

	_, err = db.Exec(todoItemsTable)
	if err != nil {
		log.Fatalf("could not create todo items table, error: %v", err)
	}
}

// AddItem ...
func (list *TodoList) AddItem(item TodoItem) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()
	stmt, err := tx.Prepare(`INSERT INTO todoItems(name, done, listId) VALUES (?, ?, ?);`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(item.Name, item.Done, list.ID)
	if err != nil {
		return err
	}
	return nil
}

// AddList ...
func AddList(list TodoList) error {
	fmt.Printf("ok")
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()
	fmt.Printf("ok")
	var stmt *sql.Stmt
	stmt, err = tx.Prepare(`INSERT INTO todoLists(name) VALUES (?);`)
	if err != nil {
		return err
	}
	fmt.Printf("ok")
	defer stmt.Close()
	_, err = stmt.Exec(list.Name)
	return err
}

// GetItems ...
func (list *TodoList) GetItems() ([]TodoItem, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Commit()
	stmt, err := tx.Prepare(`SELECT * FROM todoItems WHERE listId = ?;`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var id, listID int
	var name string
	var done bool
	rows, err := stmt.Query(list.ID)
	if err != nil {
		return nil, err
	}

	items := make([]TodoItem, 0)
	for rows.Next() {
		if err = rows.Scan(&id, &name, &done, &listID); err != nil {
			return nil, err
		}

		items = append(items, TodoItem{
			ID:     id,
			Name:   name,
			Done:   done,
			ListID: listID,
		})
	}
	return items, nil
}

// GetLists ...
func GetLists() ([]TodoList, error) {
	rows, err := db.Query(`SEELCT * FROM todoLists;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var id int
	var name string

	lists := make([]TodoList, 0)
	for rows.Next() {
		if err = rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		lists = append(lists, TodoList{
			ID: id,
			Name: name,
		})
	}
	return lists, nil
}

// GetListByID ...
func GetListByID(id int) (TodoList, error) {
	stmt, err := db.Prepare(`SELECT name FROM todoLists WHERE id = ?;`)
	if err != nil {
		return TodoList{}, err
	}
	defer stmt.Close()
	var name string
	err = stmt.QueryRow(id).Scan(&id, &name)
	if err != nil {
		return TodoList{}, err
	}
	return TodoList{ID: id, Name: name}, nil
}

// GetItemByID ...
func GetItemByID(id int) (TodoItem, error) {
	stmt, err := db.Prepare(`SELECT name, done, listId FROM todoItems WHERE id = ?;`)
	if err != nil {
		return TodoItem{}, err
	}
	defer stmt.Close()
	var listID int
	var done bool
	var name string
	err = stmt.QueryRow(id).Scan(&name, &done, &listID)
	if err != nil {
		return TodoItem{}, err
	}
	return TodoItem{ID: id, Name: name, Done: done, ListID: listID}, nil
}

// UpdateItem ...
func (list *TodoList) UpdateItem(newItem TodoItem) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()
	stmt, err := tx.Prepare(`UPDATE todoItems SET(
	    name = ?,
	    done = ?,
	    listId = ?
	) WHERE id = ?;`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(newItem.Name, newItem.Done, list.ID)
	return err
}
