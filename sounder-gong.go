package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/flosch/pongo2"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-memdb"
	"github.com/segmentio/ksuid"
)

//Song song
type Song struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Path        string `json:"path"`
}

//Songs List of Songs
type Songs struct {
	Songs []Song
}

//DB Memory Database
var DB *memdb.MemDB

func main() {

	fmt.Println("Starting application")
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	// Create the in memory database
	db, err := CreateDB()
	if err != nil {
		panic(err)
	}
	DB = db

	// Load the database if we have a state file
	songs, err := LoadDatabase()
	for _, s := range songs {

		err := s.CreateSong()
		if err != nil {
			fmt.Printf("Error loading database entry from state file with error: %s\n", err)
		}
	}

	r := mux.NewRouter()
	r.Handle("/", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(HomeHandler)))
	r.Handle("/add", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(AddHandler)))
	r.Handle("/commit", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(SaveDatabaseHandler)))
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

	songs, err := ListSongs()
	if err != nil {
		fmt.Printf("Error retrieving songs: %s", err)
	}

	template := "templates/index.html"
	//template := "templates/index.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title": "Index",
		"songs": songs,
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

	titleMessage := ""
	titleMessageError := false
	descriptionMessage := ""
	descriptionMessageError := false
	fileMessage := ""
	fileMessageError := false
	var song Song

	if r.Method == http.MethodPost {

		validate := true
		if err := r.ParseMultipartForm(10 << 20); nil != err {
			validate = false
			fileMessage = fmt.Sprintf("Error uploading file: %s", err)
			fileMessageError = true
			fmt.Println("Error Retrieving the File")
		}

		song.Title = r.FormValue("inputTitle")
		song.Description = r.FormValue("inputDescription")
		song.ID = GenUUID()

		// Validation
		if song.Title == "" {
			validate = false
			titleMessage = "Please provide a title"
			titleMessageError = true
		}

		if song.Description == "" {
			validate = false
			descriptionMessage = "Please provide a title"
			descriptionMessageError = true
		}

		// Get handler for filename, size and headers
		file, handler, fileErr := r.FormFile("inputFile")
		if fileErr != nil {

			validate = false
			fileMessage = fmt.Sprintf("Error uploading file: %s", fileErr)
			fileMessageError = true
			fmt.Println("Error Retrieving the File")
			fmt.Println(fileErr)

		}

		if validate == true {

			fmt.Printf("Uploaded File: %+v\n", handler.Filename)
			fmt.Printf("File Size: %+v\n", handler.Size)
			fmt.Printf("MIME Header: %+v\n", handler.Header)

			// Create file
			dst, err := os.Create("sounds/" + song.ID + "_" + handler.Filename)
			defer dst.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Copy the uploaded file to the created file on the filesystem
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			//fmt.Fprintf(w, "Successfully Uploaded File\n")
			err = song.CreateSong()
			if err != nil {
				fmt.Printf("Error saving song to database: %s \n", err)
			}

			song.Path = "sounds/" + handler.Filename

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return

		}

		if fileErr == nil {
			file.Close()
		}
	}

	template := "templates/add.html"
	//template := "templates/index.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":                   "Add",
		"titleMessage":            titleMessage,
		"titleMessageError":       titleMessageError,
		"descriptionMessage":      descriptionMessage,
		"descriptionMessageError": descriptionMessageError,
		"fileMessage":             fileMessage,
		"fileMessageError":        fileMessageError,
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
	go PlaySound("sounds/cartoon-birds-2_daniel-simion.wav")
}

// SaveDatabaseHandler saves the database to commit file
func SaveDatabaseHandler(w http.ResponseWriter, r *http.Request) {

	status := true
	err := CommitDatabase()
	if err != nil {
		fmt.Printf("Error Saving Database: %s", err)
		status = false
	}

	template := "templates/commit.html"
	//template := "templates/index.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":  "Commit Status",
		"status": status,
		"err":    err,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)

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

// CreateDB create new instance of in memory database
func CreateDB() (*memdb.MemDB, error) {

	// Create the DB schema
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"song": &memdb.TableSchema{
				Name: "song",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
				},
			},
		},
	}
	// Create a new data base
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, err
	}

	return db, nil
}

//CreateSong add a new song to the database
func (s *Song) CreateSong() error {

	txn := DB.Txn(true)

	err := txn.Insert("song", s)
	if err != nil {
		return err
	}

	txn.Commit()

	fmt.Printf("Added song to database ID:%s Title:%s PATH: %s\n", s.ID, s.Title, s.Path)

	return nil
}

//DeleteSong remove song from the database
func (s *Song) DeleteSong() error {

	return nil
}

//GetSongs get songs from a songs type
func (s *Songs) GetSongs() ([]Song, error) {
	var songs []Song

	for _, song := range s.Songs {
		songs = append(songs, song)
	}

	return songs, nil
}

//ListSongs Obtain a list of songs from the database
func ListSongs() ([]*Song, error) {
	var songs []*Song

	txn := DB.Txn(false)
	defer txn.Abort()

	// List all the people
	it, err := txn.Get("song", "id")
	if err != nil {
		return nil, err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		s := obj.(*Song)
		fmt.Println(s)
		songs = append(songs, s)

	}
	return songs, nil
}

//GenUUID Generate a unique identifier
func GenUUID() string {
	id := ksuid.New()
	return id.String()
}

//CommitDatabase save in memory database to a file
func CommitDatabase() error {

	songs, err := ListSongs()
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(songs)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("database/state.json", jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

//LoadDatabase loads the database from state file
func LoadDatabase() ([]Song, error) {

	var songs []Song

	fmt.Printf("Extracting json information from the state file\n")
	jsonFile, err := os.Open("database/state.json")
	if err != nil {
		return songs, err
	}

	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return songs, err
	}

	fmt.Printf("Unmarshing json data to object")
	err = json.Unmarshal(byteValue, &songs)
	if err != nil {
		return songs, err
	}

	for _, s := range songs {
		fmt.Printf("DB Load ID: %s Title: %s Path: %s\n", s.ID, s.Title, s.Path)
	}

	return songs, nil
}
