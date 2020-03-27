package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"foo.mongo-cosmosdb-test/api"
)

var crudAPI *api.API
var port string

const (
	mongoDBConnectionStringEnvVarName = "MONGODB_CONNECTION_STRING"
	dbNameEnvVarName                  = "DATABASE_NAME"
	collectionEnvVarName              = "COLLECTION_NAME"
	portEnvVarName                    = "PORT"
)

func init() {

	connectionStr := os.Getenv(mongoDBConnectionStringEnvVarName)
	if connectionStr == "" {
		log.Fatalf("missing environment variable %s", mongoDBConnectionStringEnvVarName)
	}

	dbName := os.Getenv(dbNameEnvVarName)
	if dbName == "" {
		log.Fatalf("missing environment variable %s", dbNameEnvVarName)
	}

	collectionName := os.Getenv(collectionEnvVarName)
	if collectionName == "" {
		log.Fatalf("missing environment variable %s", collectionEnvVarName)
	}

	port = os.Getenv(portEnvVarName)
	if port == "" {
		port = "8080"
	}
	crudAPI = api.NewAPI(connectionStr, dbName, collectionName)

	fmt.Println("using port", port)
}

func main() {
	r := mux.NewRouter()

	r.Methods(http.MethodPost).Path("/developers").HandlerFunc(crudAPI.Create)

	r.Methods(http.MethodGet).Path("/developers/{github}").HandlerFunc(crudAPI.Read)
	r.Methods(http.MethodGet).Path("/developers").HandlerFunc(crudAPI.ReadAll)

	r.Methods(http.MethodPut).Path("/developers").HandlerFunc(crudAPI.Update)

	r.Methods(http.MethodDelete).Path("/developers/{github}").HandlerFunc(crudAPI.Delete)

	server := http.Server{Addr: ":" + port, Handler: r}

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("could not start server %v", err)
		}
	}()

	exit := make(chan os.Signal)
	signal.Notify(exit, syscall.SIGTERM, syscall.SIGINT)
	<-exit

	fmt.Println("exit signalled")

	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		crudAPI.Close()
		cancel()
	}()
	err := server.Shutdown(c)
	if err != nil {
		log.Fatalf("failed to shutdown app %v", err)
	}
	log.Println("app shutdown")
}
