package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/DATA-DOG/golang-websocket-hub"
)

var TokenSecret = "secret-hash-code"

type Application struct {
	hub.Hub
}

func New() *Application {
	app := &Application{
		Hub: *hub.New(os.Stdout, "*"),
	}
	app.SubscriptionTokenizer = hub.HmacSha256Tokenizer(TokenSecret)
	return app
}

func (app *Application) asset(path string) {
	http.HandleFunc("/"+path, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	})
}

func (app *Application) index() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "example.html")
	})
}

func (app *Application) message() {
	http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		var msg hub.Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		to := r.URL.Query().Get("to")
		if len(to) > 0 {
			app.Mailbox <- &hub.MailMessage{
				Message:  &msg,
				Username: to,
			}
		} else {
			app.Broadcast <- &msg
		}
	})
}

func main() {
	app := New()
	app.asset("example.css")
	app.asset("example.js")
	app.message()
	app.index()

	go app.Run()
	http.Handle("/ws", app)

	log.Println("listening on: http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
