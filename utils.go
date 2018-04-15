package main

import (
	"io/ioutil"
	"log"
)

//checks if a Document exists in a single bucket using queryraw
func checkIfExist(files []Document, item string) bool {
	for _, s := range files {
		if s.Key == item {
			return true
		}
	}
	return false
}

//checkts if a Document exists in a single bucket
func checkIfExistHelper(files []Document, item Document) bool {
	for _, s := range files {
		if s == item {
			return true
		}
	}
	return false
}

//deletes specify item from files
func delete(files []Document, item string) []Document {
	for i, s := range files {
		if s.Key == item {
			return append(files[:i], files[i+1:]...)
			//return files[:i+copy(files[i:], files[i+1:])]
		}
	}
	return files
}

//write data to file after encryption
func writeToFile(data, file string) {
	//0777 Unix permition bits
	ioutil.WriteFile(file, []byte(data), 0777)
}

//read data to be encrypted
func readFromFile(file string) ([]byte, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal("could read any/all data from file", err)
	}
	return data, nil
}
