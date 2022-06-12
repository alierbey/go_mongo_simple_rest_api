package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Ctx             = context.TODO()
	BooksCollection *mongo.Collection
)

type Book struct {
	ID              primitive.ObjectID `json:"_id, omitempty" bson:"_id,omitempty"`
	Name            string             `json:"name,omitempty" bson:"name,omitempty"`
	Author          string             `json:"author,omitempty" bson:"author,omitempty"`
	PublicationDate string             `json:"publication_date,omitempty" bson:"publication_date,omitempty"`
}

func main() {

	fmt.Println("Starting the api....")

	host := "127.0.0.1"
	port := "27017"

	connectionURI := "mongodb://" + host + ":" + port + "/"
	clientOptions := options.Client().ApplyURI(connectionURI)
	client, err := mongo.Connect(Ctx, clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(Ctx, nil)

	if err != nil {
		log.Fatal(err)
	}

	db := client.Database("library")
	BooksCollection = db.Collection("book")

	r := mux.NewRouter()

	r.HandleFunc("/api/v1/books", CreateBook).Methods("POST")
	r.HandleFunc("/api/v1/books", GetBooks).Methods("GET")
	r.HandleFunc("/api/v1/books", UpdateBook).Methods("PUT")
	r.HandleFunc("/api/v1/book/{id}", DeleteBook).Methods("DELETE")
	r.HandleFunc("/api/v1/book/{id}", GetBook).Methods("GET")

	http.ListenAndServe(":8080", r)

}

func GetBooks(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("content-type", "application/json")
	var books []Book

	cursor, err := BooksCollection.Find(Ctx, bson.M{})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message":"` + err.Error() + `"}`))
		return
	}

	defer cursor.Close(Ctx)

	for cursor.Next(Ctx) {
		var book Book
		cursor.Decode(&book)
		books = append(books, book)
	}

	if err := cursor.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message":"` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(books)

}

func CreateBook(w http.ResponseWriter, r *http.Request) {

	name := r.FormValue("name")
	author := r.FormValue("author")
	publication_date := r.FormValue("publication_date")
	print(name, author, publication_date)

	var book Book
	book.Name = name
	book.Author = author
	book.PublicationDate = publication_date

	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		fmt.Println(err)
	}

	result, err := BooksCollection.InsertOne(Ctx, book)
	if err != nil {
		fmt.Println(err)
	}

	json.NewEncoder(w).Encode(result)

}

func DeleteBook(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Println(err)
	}

	filter := bson.D{{"_id", objectId}}

	result, err := BooksCollection.DeleteOne(Ctx, filter)
	if err != nil {
		fmt.Println(err)
	}

	json.NewEncoder(w).Encode(result)

}

func UpdateBook(w http.ResponseWriter, r *http.Request) {

	id := r.FormValue("id")
	name := r.FormValue("name")
	author := r.FormValue("author")
	publication_date := r.FormValue("publication_date")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Println(err)
	}

	filter := bson.D{{"_id", objectId}}
	update := bson.D{{"$set", bson.D{{"name", name}, {"author", author}, {"publication_date", publication_date}}}}

	result, err := BooksCollection.UpdateOne(Ctx, filter, update)

	if err != nil {
		fmt.Println(err)
	}
	json.NewEncoder(w).Encode(result)

}

func GetBook(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("content-type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Println(err)
	}

	var book Book

	err = BooksCollection.FindOne(Ctx, Book{ID: objectId}).Decode(&book)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message":"` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(book)

}
