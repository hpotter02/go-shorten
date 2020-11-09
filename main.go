package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/gorilla/mux"
)

type shortLink struct {
	ID          string `json:"id,omitempty"`
	Destination string `json:"destination,omitempty"`
	Name        string `json:"name,omitempty"`
}

var db *badger.DB

func main() {
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		panic(err)
	}
	defer os.Remove(dir)
	fmt.Println(dir)
	db, err = badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	fmt.Println("test")
	r := mux.NewRouter().StrictSlash(true)
	r.Handle("/", http.FileServer(http.Dir("./page"))).Methods("GET")
	r.Handle("/style.css", http.FileServer(http.Dir("./page"))).Methods("GET")
	r.Handle("/script.js", http.FileServer(http.Dir("./page"))).Methods("GET")
	r.Handle("/favicon.ico", http.FileServer(http.Dir("./page"))).Methods("GET")
	r.HandleFunc("/", handleCreateLink).Methods("POST")
	r.HandleFunc("/{id}", handleRedirect)
	http.ListenAndServe(":8800", r)
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortID := mux.Vars(r)["id"]
	u, err := getURL(shortID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
	http.Redirect(w, r, u, http.StatusMovedPermanently)
}

func handleCreateLink(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	sl := shortLink{}
	err = json.Unmarshal(b, &sl)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println(sl)
	if !isValidURL(sl.Destination) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"Invalid URL"}`)
		return
	}
	sl.ID = generateID()
	if sl.Name != "" {
		tmp, _ := getURL(sl.Name)
		if tmp != "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"Name Already Taken"}`)
			return
		}
	}

	err = setURL(sl)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	b, _ = json.Marshal(sl)
	w.Write(b)
}

func generateID() string {
	idEncdoed := make([]byte, 4)
	for {

		idNum := []byte{byte(rand.Intn(255)),
			byte(rand.Intn(255)),
			byte(rand.Intn(255))}
		idEncdoed = make([]byte, 4)
		base64.URLEncoding.Encode(idEncdoed, idNum)
		err := db.View(func(txn *badger.Txn) error {
			_, err := txn.Get(idEncdoed)
			return err
		})
		if err != nil {
			break
		}
	}
	return string(idEncdoed)
}

func setURL(s shortLink) error {
	err := db.Update(func(txn *badger.Txn) error {
		txn.Set([]byte(s.ID), []byte(s.Destination))
		if s.Name != "" {
			txn.Set([]byte(s.Name), []byte(s.Destination))
		}
		return nil
	})
	return err
}

func getURL(s string) (string, error) {
	val := make([]byte, 0)
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(s))
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}

		return err
	})
	if err != nil {
		return "", fmt.Errorf("Error Reading Key %v", err)
	}
	return string(val), nil
}

func isValidURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	if !strings.Contains(u.Scheme, "http") {
		return false
	}
	if u.Host == "" {
		return false
	}
	if u.User.String() != "" {
		return false
	}
	return true
}
