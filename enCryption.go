package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"log"
)

//generates a key based on the bucketname and filename
func generateKey(filename, bucketname string) string {
	key := make([]byte, len(filename)+len(bucketname))
	_, err := rand.Read(key)
	if err != nil {
		log.Println("cannot create any key", err)
	}
	return string(key)
}

//encrypt file's content
func encryption(plaintext string, keystring string) (string, error) {

	key := []byte(keystring)

	Block, err := aes.NewCipher(key)
	if err != nil {
		log.Println("could not create any cipher block", err)
	}

	text := []byte(plaintext)
	//empty array of 16 + length of plaintext including the iv
	ciphertext := make([]byte, aes.BlockSize+len(text))

	//iv slice of 16 byte for randomization
	iv := ciphertext[:aes.BlockSize]

	//write 16 rand byte to fill iv
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Fatal("Error when initializing iv ", err)
	}

	//return encrypted stream
	stream := cipher.NewCFBEncrypter(Block, iv)

	//encrypt bytes from plaintext to cipher text
	stream.XORKeyStream(ciphertext[aes.BlockSize:], text)

	return string(ciphertext), nil
}

//decrypt file's content
func decrypt(cipherstring string, keystring string) (string, error) {

	key := []byte(keystring)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println("could not create any cipher block", err)
	}

	ciphertext := []byte(cipherstring)

	//always test if the ciphertext len is long enough
	if len(ciphertext) < aes.BlockSize {
		log.Fatal("text to short ", err)
	}

	//get the 16 byte iv
	iv := ciphertext[:aes.BlockSize]

	//remove the iv from the ciphertext
	ciphertext = ciphertext[aes.BlockSize:]

	//return a decrypted stream
	stream := cipher.NewCFBDecrypter(block, iv)

	//Decrypt byte from ciphertext
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}
