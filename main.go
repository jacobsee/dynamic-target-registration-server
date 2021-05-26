package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	bolt "go.etcd.io/bbolt"
)

var db *bolt.DB

func main() {
	databasePath := os.Getenv("DB_FILE")
	if len(databasePath) == 0 {
		databasePath = "targets.db"
	}
	var err error
	db, err = bolt.Open(databasePath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/register", register)
	http.HandleFunc("/unregister", unregister)
	http.HandleFunc("/list", list)

	log.Println("Database initialization complete - starting HTTP server now.")

	http.ListenAndServe(":8081", nil)
}

func register(w http.ResponseWriter, r *http.Request) {
	// Handle auth
	auth := r.Header.Get("Authorization")
	expectedAuth := os.Getenv("AUTH_TOKEN")
	if len(auth) == 0 || auth != expectedAuth {
		log.Println("Could not authorize registration request (Authorization header is required)")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Parse request
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

	// Make sure the bucket (target type) exists
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

	// Store target info
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
	// Handle auth
	auth := r.Header.Get("Authorization")
	expectedAuth := os.Getenv("AUTH_TOKEN")
	if len(auth) == 0 || auth != expectedAuth {
		log.Println("Could not authorize registration request (Authorization header is required)")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Parse request
	err := r.ParseForm()
	if err != nil {
		panic(err)
	}
	kind := r.Form.Get("kind")
	key := r.Form.Get("key")
	if kind == "" || key == "" {
		log.Println("Unregister request missing required parameters")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Remove target info
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(kind))
		b.Delete([]byte(key))
		if err != nil {
			return fmt.Errorf("Could not unregister target: %s", err)
		}
		return nil
	})
}

func list(w http.ResponseWriter, r *http.Request) {
	// Handle auth
	auth := r.Header.Get("Authorization")
	expectedAuth := os.Getenv("AUTH_TOKEN")
	if len(auth) == 0 || auth != expectedAuth {
		log.Println("Could not authorize registration request (Authorization header is required)")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Parse request
	kind, ok := r.URL.Query()["kind"]
	if !ok || len(kind[0]) == 0 {
		log.Println("List request missing required parameters")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Fetch requested data
	answer := map[string]string{}
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(kind[0]))
		if b == nil {
			return fmt.Errorf("Received request to list a bucket that does not exist: %s", kind[0])
		}
		b.ForEach(func(key, value []byte) error {
			answer[string(key)] = string(value)
			return nil
		})
		return nil
	})
	if err != nil {
		log.Printf("Error retrieving contents of bucket: %s\n", kind[0])
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
