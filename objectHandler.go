package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/minio/minio-go"
	"github.com/minio/minio-go/pkg/encrypt"
)

//Document is used to specify any object into a bucket.
type Document struct {
	Key string
	URL string
}

//ListObjects is a linked list of Documents
//type ListObjects struct{ list.Element }

//Objects are all available objects in a single bucket
var Objects = make([]Document, 1000)

//DocuHandler is the an instance of minio client.
type DocuHandler struct {
	minioclient *minio.Client
	//TLSConfig is the configuration for the tls connection
	TLSConfig *tls.Config
}

//IsSecure checks wether a connection is http(insecure) or https(secure)
func IsSecure(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}
	return u.Scheme == "https"
}

//FindHost search for available host with a given url
func FindHost(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}
	return u.Host
}

//ListOneObject uses to list just one object from one specified bucket.
func (dok DocuHandler) ListOneObject(w http.ResponseWriter, r *http.Request) {

	objectName := r.FormValue("object")

	err := dok.minioclient.FGetObject(*BucketName, objectName, "objectName", minio.GetObjectOptions{})
	if err != nil {
		http.Error(w, objectName, http.StatusInternalServerError)
		log.Println("error when searching the item ", err)
		return
	}

	dokListJSON, err := json.Marshal(objectName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(dokListJSON)
}

//ListObjectsHandler lists objects in a bucket.
func (dok DocuHandler) ListObjectsHandler(w http.ResponseWriter, r *http.Request) {

	doneCh := make(chan struct{})

	var docu []Document

	var isrecursive = true

	var firstObjectPresigned bool

	defer close(doneCh)

	for objectInfo := range dok.minioclient.ListObjects(*BucketName, "", isrecursive, doneCh) {
		if objectInfo.Err != nil {
			http.Error(w, objectInfo.Err.Error(), http.StatusInternalServerError)
			log.Println(objectInfo.Err)
			return
		}
		objName := objectInfo.Key

		dokElement := Document{
			Key: objName,
		}
		if !firstObjectPresigned {
			expiring := 400 * time.Second
			firstObjectPresigned, err := dok.minioclient.PresignedGetObject(*BucketName, objName, expiring, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			dokElement.URL = firstObjectPresigned.String()
		}

		docu = append(docu, dokElement)

		if checkIfExistHelper(Objects, dokElement) {
			return
		}
		Objects = append(Objects, dokElement)
		//fmt.Println("elements of Objects ", docu[1])

	}

	dokListJSON, err := json.Marshal(docu)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(dokListJSON)
}

//GetPresignedURLHandler generates access URL for object in a bucket.
func (dok DocuHandler) GetPresignedURLHandler(w http.ResponseWriter, r *http.Request) {
	objectName := r.FormValue("object")
	expiring := 400 * time.Second
	if objectName == " " {
		http.Error(w, "no object found with this name ", http.StatusInternalServerError)
		return
	}
	firstObjectPresigned, err := dok.minioclient.PresignedGetObject(*BucketName, objectName, expiring, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Write([]byte(firstObjectPresigned.String()))
}

//RemoveObjectHandler deletes an object from a bucket.
func (dok DocuHandler) RemoveObjectHandler(w http.ResponseWriter, r *http.Request) {

	objectsCh := make(chan string)

	objectName := r.FormValue("object")

	secure := false

	go func() {
		defer close(objectsCh)

		//all objects from a bucket-name
		for objectInfo := range dok.minioclient.ListObjects(*BucketName, "", secure, nil) {
			if objectInfo.Err != nil {
				http.Error(w, objectInfo.Err.Error(), http.StatusInternalServerError)
			}
			objectsCh <- objectInfo.Key
		}
	}()

	rErr := dok.minioclient.RemoveObject(*BucketName, objectName)

	//checkIfExist if the given object is available
	if checkIfExist(Objects, objectName) == false || objectName == "" {
		http.Error(w, "object does not exist ", http.StatusInternalServerError)
		return
	}

	if rErr != nil {
		http.Error(w, "error when deleting ", http.StatusInternalServerError)
	}

	Objects = delete(Objects, objectName)

	log.Println("object successfully deleted ", objectName)
}

//PostObjectHandler updates an existing file to a bucket.
func (dok DocuHandler) PostObjectHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("object")
	if err != nil {
		log.Println("error when reading file ", err)
	}

	defer file.Close()
	fmt.Fprintf(w, "%v", handler.Header)

	f, err := os.OpenFile(handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	content, _ := readFromFile(f.Name())
	//Generate key for encryption
	keyEncrypt := generateKey(string(content), *BucketName)
	//log.Println("key generated ", string(keyEncrypt))

	writeToFile(keyEncrypt, f.Name())

	defer f.Close()
	io.Copy(f, file)

	object, err := dok.minioclient.FPutObject(*BucketName, handler.Filename, handler.Filename, minio.PutObjectOptions{ContentType: "application/octet-stream"})

	dokElement := Document{
		Key: handler.Filename,
	}

	if checkIfExistHelper(Objects, dokElement) {
		http.Error(w, "given object already exist ", http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, "error when updloading file ", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	Objects = append(Objects, dokElement)

	log.Println("object successfully uploaded ", object, dokElement.Key)
}

//PutEncryptedObject uses to encrypt an object with SSE-C
func (dok DocuHandler) PutEncryptedObject(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("object")
	if err != nil {
		log.Println("error when reading file ", err)
	}

	defer file.Close()

	fmt.Fprintf(w, "%v", handler.Header)
	f, err := os.Open(handler.Filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer f.Close()

	fstat, err := f.Stat()
	if err != nil {
		log.Fatal("errorr when retreiving the file infos ", err)
	}
	io.Copy(f, file)

	encryption := encrypt.DefaultPBKDF([]byte(secretKey), []byte(*BucketName+handler.Filename))
	fmt.Printf("encryption key %s ", encryption)

	n, err := dok.minioclient.PutObject(*BucketName, handler.Filename, f, fstat.Size(), minio.PutObjectOptions{ServerSideEncryption: encryption})
	if err != nil {
		log.Fatal("error when uploading encrypted file ", err)
	}
	log.Println("Succesfull uploaded ", handler.Filename, "of size", n, "about file", fstat)

}
