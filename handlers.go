package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	UploadBufferSize = 1024 * 1024 * 4 // 4MB
)

var client = createClient()

func handlePing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong\n")
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	var err error
	host := os.Getenv("GYAZO_HOST") || "http://localhost:3000"

	err = r.ParseMultipartForm(UploadBufferSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m := r.MultipartForm
	files := m.File["imagedata"]

	for i, _ := range files {
		file, err := files[i].Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		hash, err := upload(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%s/%s.png", host, hash)
		return
	}

	http.Error(w, "imagedata is not specified", http.StatusBadRequest)
}

func upload(file multipart.File) (string, error) {
	hasher := md5.New()
	io.Copy(hasher, file)
	hash := hex.EncodeToString(hasher.Sum(nil))
	filename := fmt.Sprintf("%s.png", hash)

	bucketName := os.Getenv("S3_BUCKET_NAME")
	param := &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &filename,
		Body:   file,
	}
	_, err := client.PutObject(param)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func createClient() *s3.S3 {
	return s3.New(session.New(), createConfig())
}

func createConfig() *aws.Config {
	cred := getCred()
	config := aws.NewConfig().WithCredentials(cred)
	config.Region = aws.String("ap-northeast-1")
	return config
}

func getCred() *credentials.Credentials {
	var cred *credentials.Credentials
	envCredential := credentials.NewEnvCredentials()
	envCredValue, err := envCredential.Get()
	if err != nil {
		log.Println(err)
	}
	sharedCredential := credentials.NewSharedCredentials("", "default")
	sharedCredValue, err := sharedCredential.Get()
	if err != nil {
		log.Println(err)
	}
	if envCredValue.AccessKeyID != "" && envCredValue.SecretAccessKey != "" {
		cred = envCredential
	} else if sharedCredValue.AccessKeyID != "" && sharedCredValue.SecretAccessKey != "" {
		cred = sharedCredential
	} else {
		log.Panicln("No credentials found")
	}
	return cred
}
