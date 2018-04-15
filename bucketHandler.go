package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

//Buckets represents a slice for all available buckets.
var Buckets []BucketInfo

//BucketInfo represents infos about a bucket
type BucketInfo struct {
	Name         string
	CreationDate time.Time
}

//PostNewBucket creates a new bucket
func (dok DocuHandler) PostNewBucket(w http.ResponseWriter, r *http.Request) {

	bucket := BucketInfo{
		Name:         *BucketName,
		CreationDate: time.Now(),
	}

	exist, erOr := dok.minioclient.BucketExists(bucket.Name)
	if erOr != nil && exist {
		http.Error(w, "bucket with given name does already exist", http.StatusInternalServerError)
		log.Println(erOr)
		return
	}

	err := dok.minioclient.MakeBucket(bucket.Name, "us-east-1")

	if err != nil {
		http.Error(w, "error when creating bucket", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	log.Println("bucket succesfully created at:", time.Now())

}

//ListBuckets lists all available buckets
func (dok DocuHandler) ListBuckets(w http.ResponseWriter, r *http.Request) {

	var bukets []BucketInfo

	buckets, err := dok.minioclient.ListBuckets()

	if err != nil {
		http.Error(w, "error when listing buckets", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	for _, bucket := range buckets {
		if bucket.Name == "" {
			http.Error(w, "error bucket's name empty", http.StatusInternalServerError)
			return
		}

		newBucket := BucketInfo{
			Name: bucket.Name,
		}
		bukets = append(bukets, newBucket)
	}

	for _, b := range bukets {
		newBucket := BucketInfo{
			Name: b.Name,
		}
		Buckets = append(Buckets, newBucket)
	}

	bucketList, err := json.Marshal(bukets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(bucketList)
}

//RemoveBucket deletes a specified bucket's name
func (dok DocuHandler) RemoveBucket(w http.ResponseWriter, r *http.Request) {

	bucketN := r.URL.Query().Get("bucket")
	objectsCh := make(chan string)

	go func() {
		defer close(objectsCh)

		buckets, err := dok.minioclient.ListBuckets()

		if err != nil {
			http.Error(w, "error when listing buckets", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		for _, bucket := range buckets {
			if bucket.Name == "" {
				http.Error(w, "error bucket's name empty", http.StatusInternalServerError)
				return
			}
			objectsCh <- bucket.Name
		}
	}()

	err := dok.minioclient.RemoveBucket(bucketN)
	if err != nil {
		http.Error(w, "error could not delete any bucket", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	log.Println("object succesfully deleted ", bucketN)
}
