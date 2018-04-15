package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/minio/minio-go"
)

var (
	accessKey = "C8NODTMLRKRBB4665CV9"
	secretKey = "a3xpzQFu4rKRC/pm0Ky9ciyhBvXMb/Go3cfETNH2"
)

var (
	//BucketName represents bucket of minio
	BucketName = flag.String("B", "docusafe", "bucket name to be used for documents assets.")
	//Endpoint represents the endpoint of minio
	Endpoint = flag.String("E", "http://172.16.122.216:9000", "custom endpoint.")
	//CaFile root certificate file
	CaFile = flag.String("CA", "RootCertificate", "A PEM encoded CA's certificate file.")
	//KeyFile private certificate
	KeyFile = flag.String("key", "keyfile", "A PEM encoded private key file.")
)

func main() {

	flag.Parse()

	flag.PrintDefaults()

	if BucketName == nil {
		log.Fatal("bucketname might to be empty.", BucketName)
	}

	/* //Load CA cert
	caCert, err := ioutil.ReadFile(*CaFile)
	if err != nil {
		log.Fatal("error when reading root certificate ", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert) */

	//create new instance of minioClient
	minioClient, err := minio.New(FindHost(*Endpoint), accessKey, secretKey, IsSecure(*Endpoint))

	if err != nil {
		log.Fatalln("could not find any server", err)
	}

	fileStorage := DocuHandler{
		minioclient: minioClient,
	}

	http.HandleFunc("/getpresignedlist/v1", fileStorage.GetPresignedURLHandler)
	http.HandleFunc("/listfiles/v1", fileStorage.ListObjectsHandler)
	http.HandleFunc("/postnewfile/v1", fileStorage.PostObjectHandler)
	http.HandleFunc("/removefile/v1", fileStorage.RemoveObjectHandler)
	http.HandleFunc("/list1file/v1", fileStorage.ListOneObject)
	http.HandleFunc("/postnewbucket/v1", fileStorage.PostNewBucket)
	http.HandleFunc("/listbuckets/v1", fileStorage.ListBuckets)
	http.HandleFunc("/deletebucket/v1", fileStorage.RemoveBucket)
	//http.HandleFunc("/postsecuredfile/v1", fileStorage.EncryptObject)
	http.HandleFunc("/putencryptedfile/v1", fileStorage.PutEncryptedObject)

	log.Println("dokusafe started successfully at : http://localhost:8443")

	/* 	err = http.ListenAndServeTLS(":8443", "public.crt", "private.key", nil)
	   	if err != nil {
	   		log.Println("ListenAndServe", err)
	   	} */

	http.ListenAndServe(":8080", nil)
}
