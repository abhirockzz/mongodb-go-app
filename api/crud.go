package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/abhirockzz/mongodb-go-app/model"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// API represents a handle to a MongoDB cluster
type API struct {
	Connection     *mongo.Client
	DBName         string
	CollectionName string
}

// NewAPI returns an a handle to API
func NewAPI(connectionString, db, collection string) *API {
	return &API{Connection: getConnection(connectionString), DBName: db, CollectionName: collection}
}

// Create persists a new developer profile
func (a *API) Create(rw http.ResponseWriter, req *http.Request) {

	coll := a.Connection.Database(a.DBName).Collection(a.CollectionName)

	var dev model.Developer
	json.NewDecoder(req.Body).Decode(&dev)

	ctx := context.Background()
	res, err := coll.InsertOne(ctx, dev)

	if err != nil {
		failed := "failed to create developer profile"
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(rw, failed)
		return
	}

	id := res.InsertedID.(primitive.ObjectID)
	fmt.Println("inserted details. ID =", id.String())
	rw.WriteHeader(http.StatusCreated)
}

const githubIDAttribute = "github_id"

// Read fetches developer profile given the github ID
func (a *API) Read(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	githubID := vars["github"]

	coll := a.Connection.Database(a.DBName).Collection(a.CollectionName)
	r := coll.FindOne(context.Background(), bson.M{githubIDAttribute: githubID})

	var failed string
	if r.Err() != nil && r.Err() == mongo.ErrNoDocuments {
		failed = "developer profile with github ID " + githubID + " does not exist"
		rw.WriteHeader(http.StatusNotFound)
		log.Println(failed, r.Err())
		fmt.Fprintln(rw, failed)
		return
	} else if r.Err() != nil {
		failed = "operation failed"
		log.Println(r.Err())
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(rw, failed)
		return
	}
	var p model.Developer
	r.Decode(&p)
	json.NewEncoder(rw).Encode(&p)
}

// ReadAll fetches all available dveloper profile info
func (a *API) ReadAll(rw http.ResponseWriter, req *http.Request) {
	coll := a.Connection.Database(a.DBName).Collection(a.CollectionName)
	r, err := coll.Find(context.Background(), bson.D{})

	var failed string
	if err != nil {
		failed = "operation failed"
		log.Println(r.Err())
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(rw, failed)
		return
	}
	devs := []model.Developer{}
	err = r.All(context.Background(), &devs)
	if err != nil {
		failed = "operation failed"
		log.Println(r.Err())
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(rw, failed)
		return
	}

	json.NewEncoder(rw).Encode(&devs)
}

// Update updates developer profile
func (a *API) Update(rw http.ResponseWriter, req *http.Request) {
	var dev model.Developer
	json.NewDecoder(req.Body).Decode(&dev)

	coll := a.Connection.Database(a.DBName).Collection(a.CollectionName)
	githubID := dev.GithubHandle
	r := coll.FindOneAndReplace(context.Background(), bson.M{githubIDAttribute: githubID}, &dev)

	var failed string
	if r.Err() != nil && r.Err() == mongo.ErrNoDocuments {
		failed = "developer profile with github ID " + githubID + " does not exist"
		log.Println(r.Err())
		rw.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(rw, failed)
		return
	} else if r.Err() != nil {
		failed = "operation failed"
		log.Println(r.Err())
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(rw, failed)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

// Delete removes developer profile
func (a *API) Delete(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	githubID := vars["github"]

	coll := a.Connection.Database(a.DBName).Collection(a.CollectionName)
	r := coll.FindOneAndDelete(context.Background(), bson.M{githubIDAttribute: githubID})

	var failed string
	if r.Err() != nil && r.Err() == mongo.ErrNoDocuments {
		failed = "developer profile with github ID " + githubID + " does not exist"
		log.Println(r.Err())
		rw.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(rw, failed)
		return
	} else if r.Err() != nil {
		failed = "operation failed"
		log.Println(failed, r.Err())
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(rw, failed)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

// Close disconnects from MongoDB
func (a API) Close() {
	err := a.Connection.Disconnect(context.Background())
	if err != nil {
		log.Println("failed to disconnect")
		return
	}
	log.Println("disconnected from MongoDB")
}

func getConnection(connectionStr string) *mongo.Client {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	clientOptions := options.Client().ApplyURI(connectionStr)
	c, err := mongo.NewClient(clientOptions)

	err = c.Connect(ctx)

	if err != nil {
		log.Fatalf("unable to connect %v", err)
	}
	err = c.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("ping failed %v", err)
	}

	log.Println("connected to MongoDB")
	return c
}
