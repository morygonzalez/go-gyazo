package main

import (
	"fmt"
	"net/http"
	"os"
	"syscall"
)

var port string

func init() {
	pidFilePath := os.Getenv("PIDFILE") || "go-gyazo.pid"
	if ferr := os.Remove(pidFilePath); ferr != nil {
		if !os.IsNotExist(ferr) {
			panic(ferr.Error())
		}
	}
	pidf, perr := os.OpenFile(pidFilePath, os.O_EXCL|os.O_CREATE|os.O_WRONLY, 0666)

	if perr != nil {
		panic(perr.Error())
	}
	if _, err := fmt.Fprint(pidf, syscall.Getpid()); err != nil {
		panic(err.Error())
	}
	pidf.Close()
}

func main() {
	envPort := os.Getenv("PORT")
	if envPort != "" {
		port = envPort
	} else {
		port = "3000"
	}

	addr := fmt.Sprintf("0.0.0.0:%s", port)

	handle("/ping", handlePing)
	handle("/upload.cgi", routeByMethods(methodHandlerMap{"POST": handleUpload}))

	fmt.Printf("Listening on %s...\n", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}
