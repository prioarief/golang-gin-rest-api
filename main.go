package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type Todo struct {
	ID     int    `json:"id"`
	Task   string `json:"task"`
	Status string `json:"status"`
}

func main() {
	r := gin.Default()

	r.GET("/todos", getTodos)
	r.GET("/todos/:id", getTodo)
	r.POST("/todos", createTodo)
	r.PUT("/todos/:id", updateTodo)
	r.DELETE("/todos/:id", deleteTodo)

	if err := r.Run(":8089"); err != nil {
		log.Fatal(err)
	}
}

func getTodos(c *gin.Context) {
	db := getDB()
	defer db.Close()

	rows, err := db.Query("SELECT id, task, status FROM todos")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	todos := make([]Todo, 0)
	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.ID, &todo.Task, &todo.Status)
		if err != nil {
			log.Fatal(err)
		}
		todos = append(todos, todo)
	}

	c.JSON(http.StatusOK, todos)
}

func getTodo(c *gin.Context) {
	db := getDB()
	defer db.Close()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Fatal(err)
	}

	var todo Todo
	err = db.QueryRow("SELECT id, task, status FROM todos WHERE id = ?", id).Scan(&todo.ID, &todo.Task, &todo.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		} else {
			log.Fatal(err)
		}
		return
	}

	c.JSON(http.StatusOK, todo)
}

func createTodo(c *gin.Context) {
	db := getDB()
	defer db.Close()

	var todo Todo
	if err := c.BindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	_, err := db.Exec("INSERT INTO todos (task, status) VALUES (?, ?)", todo.Task, todo.Status)
	if err != nil {
		log.Fatal(err)
	}

	// todo.ID, _ = result.LastInsertId()

	c.JSON(http.StatusCreated, todo)
}

func updateTodo(c *gin.Context) {
	db := getDB()
	defer db.Close()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Fatal(err)
	}

	var todo Todo
	if err := c.BindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	result, err := db.Exec("UPDATE todos SET task = ?, status = ? WHERE id = ?", todo.Task, todo.Status, id)
	if err != nil {
		log.Fatal(err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	c.Status(http.StatusOK)
}

func deleteTodo(c *gin.Context) {
	db := getDB()
	defer db.Close()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Fatal(err)
	}

	result, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		log.Fatal(err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	c.Status(http.StatusOK)
}

func getDB() *sql.DB {
	db, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/golang")
	if err != nil {
		log.Fatal(err)
	}

	return db
}
