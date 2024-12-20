package main

import (
	"context"
	"log"
	"log-service/data"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	httpPort = "8080"
	rpcPort  = "5001"
	mongoURL = "mongodb://mongo:27017"
	gRPCPort = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	// connect to mongo
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}

	client = mongoClient

	// create a context in order to disconnect from mongo
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// disconnect from mongo
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	// Register the RPC server
	err = rpc.Register(&RPCServer{Models: &app.Models})
	if err != nil {
		log.Panic(err)
	}
	go app.rpcListen()
	// start the server
	app.serve()
}

func (app *Config) serve() {
	server := &http.Server{
		Addr:    ":" + httpPort,
		Handler: app.routes(),
	}

	log.Printf("Starting server on port %s", httpPort)
	if err := server.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}

func (app *Config) rpcListen() error {
	log.Println("Starting RPC server on port", rpcPort)
	listen, err := net.Listen("tcp", ":"+rpcPort)
	if err != nil {
		log.Println("Error starting RPC server:", err)
		return err
	}

	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			log.Println("Error accepting connection on RPC server:", err)
			continue
		}
		go rpc.ServeConn(rpcConn)
	}
}

func connectToMongo() (*mongo.Client, error) {
	// create connection options
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: "root",
		Password: "password",
	})

	// connect to mongo
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Printf("Error connecting to mongo: %v", err)
		return nil, err
	}

	log.Println("Connected to mongo")
	return client, nil
}
