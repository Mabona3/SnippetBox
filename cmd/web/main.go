package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"snippetbox.mabona3.net/internal/models"
)

type application struct {
	infoLog       *log.Logger
	errorLog      *log.Logger
	snippets      *models.SnippetModel
	templateCache map[string]*template.Template
	formDecoder   *schema.Decoder
	Store         *sessions.CookieStore
}

func main() {
	var addr string
	var dsn string
	var store sessions.CookieStore
	getVars(&dsn, &addr, &store)

	errLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(dsn)
	if err != nil {
		errLog.Fatal(err)
	}
	defer db.Close()

	newtemplateCache, err := newTemplateCache()
	if err != nil {
		errLog.Fatal(err)
	}

	formDecoder := schema.NewDecoder()

	a := application{
		infoLog:       log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		errorLog:      errLog,
		snippets:      &models.SnippetModel{DB: db},
		templateCache: newtemplateCache,
		formDecoder:   formDecoder,
		Store:         &store,
	}

	srv := &http.Server{
		Addr:     addr,
		ErrorLog: a.errorLog,
		Handler:  a.routes(),
	}

	a.infoLog.Printf("Starting server on %s", addr)
	err = srv.ListenAndServe()
	a.errorLog.Fatal(err)
}

func getVars(dsn *string, addr *string, store *sessions.CookieStore) {
	godotenv.Load(".env")

	store = sessions.NewCookieStore([]byte(os.Getenv("SECRET_KEY")))

	*addr = *flag.String("addr", ":"+os.Getenv("PORT"), "HTTP network address")
	*dsn = *flag.String("dsn", os.Getenv("DSN"), "MySQL data source name")
	flag.Parse()
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
