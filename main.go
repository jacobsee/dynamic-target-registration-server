package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	bolt "go.etcd.io/bbolt"
)

var db *bolt.DB

func main() {
	var err error
	db, err = bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/register", register)
	http.HandleFunc("/unregister", unregister)
	http.HandleFunc("/list", list)
	http.ListenAndServe(":8081", nil)
}

func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		panic(err)
	}
	kind := r.Form.Get("kind")
	key := r.Form.Get("key")
	data := r.Form.Get("data")
	if kind == "" || key == "" || data == "" {
		log.Println("Register request missing required parameters")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("Will add target of kind %s", kind)

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(kind))
		if err != nil {
			return fmt.Errorf("Could not create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(kind))
		err := b.Put([]byte(key), []byte(data))
		return err
	})
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func unregister(w http.ResponseWriter, r *http.Request) {

}

func list(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		panic(err)
	}
	kind := r.Form.Get("kind")
	if kind == "" {
		log.Println("List request missing required parameters")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	answer := map[string]string{}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(kind))
		if b == nil {
			return fmt.Errorf("Received request to list a bucket that does not exist: %s", kind)
		}
		b.ForEach(func(key, value []byte) error {
			answer[string(key)] = string(value)
			return nil
		})
		return nil
	})
	if err != nil {
		// log.Fatal(err)
		log.Printf("Error retrieving contents of bucket: %s\n", kind)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	output, err := json.Marshal(answer)
	if err != nil {
		log.Printf("Failed to marshal bucket content to JSON")
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(output)
}
