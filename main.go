package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html"
)

type todo struct {
	Item string
}

func main() {
	pgUser := or(or(os.Getenv("DB_USER"), os.Getenv("PGUSER")), "postgres")
	pgPassword := or(os.Getenv("DB_PASSWORD"), os.Getenv("PGPASSWORD"))
	pgHost := or(or(os.Getenv("DB_HOST"), os.Getenv("PGHOST")), "localhost:5432")
	pgSSLMode := or(or(os.Getenv("DB_SSL_MODE"), os.Getenv("PGSSLMODE")), "require")
	dbName := or(or(os.Getenv("DB_NAME"), os.Getenv("DBNAME")), "mydb")

	connStr := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=%s", pgUser, pgPassword, pgHost, dbName, pgSSLMode)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	//checked by kubernetes to see if the pod is ready to receive traffic
	app.Get("/healthz", func(c *fiber.Ctx) error {
		fmt.Println("healthcheck")
		err := db.Ping()
		if err != nil {
			c.SendString(err.Error())
		}
		return err
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return getTodos(c, db)
	})

	app.Post("/", func(c *fiber.Ctx) error {
		return newTodo(c, db)
	})

	app.Delete("/delete", func(c *fiber.Ctx) error {
		return deleteTodo(c, db)
	})

	port := or(os.Getenv("PORT"), "8080")
	app.Static("/", "./public")
	app.Use(logger.New())

	//we need to keep re-trying until successful, but don't want to block
	//the api form starting, so we kick off a go-routine
	go func() {
		x := 0
		for {
			log.Println("Attempting to connect to DB")

			if err := initDB(db); err == nil {
				break
			}

			if x > 60 {
				log.Printf("Retried %d times, exiting\n", x)
				log.Fatal(err)
			}

			log.Printf("Failed to connect to DB, retry attempt %d/60. Err: %v\n", x, err)
			time.Sleep(time.Second)
			x++
		}
	}()
	log.Println(app.Listen(fmt.Sprintf(":%v", port)))
}

func getTodos(c *fiber.Ctx, db *sql.DB) error {
	var res string
	var todos []string
	rows, err := db.Query("SELECT * FROM todos")
	if err != nil {
		log.Fatalln(err)
		c.JSON("An error occured")
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&res)
		todos = append(todos, res)
	}
	return c.Render("index", fiber.Map{
		"Todos":      todos,
		"Enterprise": os.Getenv("ENTERPRISE"),
		"PCI": os.Getenv("CONTAINS_CC"),
	})
}

func newTodo(c *fiber.Ctx, db *sql.DB) error {
	newTodo := todo{}
	if err := c.BodyParser(&newTodo); err != nil {
		log.Printf("An error occured: %v", err)
		return c.SendString(err.Error())
	}
	fmt.Printf("Creating a new To Do: %q\n", newTodo)
	if newTodo.Item != "" {
		_, err := db.Exec("INSERT into todos VALUES ($1)", newTodo.Item)
		if err != nil {
			log.Fatalf("An error occured while executing query: %v", err)
		}
	}

	return c.Redirect("/")
}

func deleteTodo(c *fiber.Ctx, db *sql.DB) error {
	todoToDelete := c.Query("item")
	db.Exec("DELETE from todos WHERE item=$1", todoToDelete)
	fmt.Printf("Deleting To Do: %q\n", todoToDelete)
	return c.SendString("deleted")
}

func initDB(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS todos (item text)")
	return err
}

func or(a string, b string) string {
	if a == "" {
		return b
	}
	return a
}
