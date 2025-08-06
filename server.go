package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This thing is working! Your url now btw: %s", r.URL.Path)
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := getTasks()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting tasks: %v", err), http.StatusInternalServerError)
		return
	}
	if len(tasks) == 0 {
		fmt.Fprintln(w, "No tasks found.")
		return
	} else {
		fmt.Fprintln(w, "Tasks:")
		for _, task := range tasks {
			fmt.Fprintf(w, "- %s\n", task)
		}
	}
}

func connectToDatabase() *pgx.Conn {
	connStr := "user=akon password=lislignaomo host=localhost port=5434 dbname=postgres sslmode=disable"
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	err = conn.Ping(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ping failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully connected to PostgreSQL!")
	return conn
}

func createTableIfNotExists(conn *pgx.Conn) {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS tasks (
		id SERIAL PRIMARY KEY,
		description TEXT NOT NULL UNIQUE
	);`

	_, err := conn.Exec(context.Background(), createTableSQL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating table: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Table 'tasks' created or already exists.")
}

func insertTask(conn *pgx.Conn, description string) {
	insertSQL := "INSERT INTO tasks (description) VALUES ($1)"
	_, err := conn.Exec(context.Background(), insertSQL, description)
	if err != nil {
		uniqueViolationCode := "23505"
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			fmt.Println("Task already exists, not inserted again.")
		} else {
			fmt.Fprintf(os.Stderr, "Error inserting task: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("New task inserted successfully!")
	}
}

func getTasks() ([]string, error) {
	conn := connectToDatabase()
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "SELECT description FROM tasks")
	if err != nil {
		return nil, fmt.Errorf("error querying tasks: %w", err)
	}
	defer rows.Close()

	var tasks []string
	for rows.Next() {
		var description string
		if err := rows.Scan(&description); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		tasks = append(tasks, description)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating over rows: %w", err)
	}

	return tasks, nil
}

func main() {
	conn := connectToDatabase()
	defer conn.Close(context.Background())

	createTableIfNotExists(conn)
	insertTask(conn, "Learn Go with the Google course")

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/tasks", tasksHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server: ", err)
	}

	fmt.Println("Server listening on port 8080...")
}
