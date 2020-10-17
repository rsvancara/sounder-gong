package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/flosch/pongo2"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

func main() {

	fmt.Println("Starting application")
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	r := mux.NewRouter()
	r.Handle("/", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(HomeHandler)))
	r.Handle("/add", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(AddHandler)))
	r.Handle("/delete/{soundid}", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(DeleteHandler))) 
	r.Handle("/play/{soundid}", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(PlaySoundHandler)))

	dir := "./static" 
        r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))

	http.Handle("/", r)

	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}

// HomeHandler home page handler that displays list of songs
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)

	template := "templates/index.html"
	//template := "templates/index.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":     "Index",
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)

}

// AddHandler handler for adding songs
func AddHandler(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)

	template := "templates/add.html"
	//template := "templates/index.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":     "Add",
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)

}

// DeleteHandler handler for deleting songs
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "delete")
}

// PlaySoundHandler handler for playing songs 
func PlaySoundHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "It Works")

	// Play sound in a go routine so it does not block
	go PlaySound("cartoon-birds-2_daniel-simion.wav")
}

// PlaySound plays the sound specified by the soundfile parameter that must be a valid path to a sound file
func PlaySound(soundfile string) {

	// This is possiby the most horrific way to play a sound, but it works 
	cmd := exec.Command("/usr/bin/aplay", soundfile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%s\n", out)
}
