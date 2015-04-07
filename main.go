package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	host   *string
	dbname *string
	col    *string
	out    *string
	fields *string
	keys   = StringArray{}
)

func init() {
	host = flag.String("H", "localhost:27017", "The host and port of the mongod instance you wish to connect to")
	dbname = flag.String("db", "testDB", "The output file that will be written to")
	//user = flag.String("user", "", "The user you wish to authenticate with")
	//pass = flag.String("pass", "", "The pass you wish to authenticate with")
	col = flag.String("c", "testCol", "The collection you wish to output")
	flag.Var(&keys, "f", "fields to output")
	// query := flag.String("query", "collection.csv", "The output file that will be written to")
	out = flag.String("o", "collection.csv", "The output file that will be written to")
}

var flattened = make(map[string]interface{})

func flatten(input bson.M, lkey string, flattened *map[string]interface{}) {
	for rkey, value := range input {
		key := lkey + rkey
		if _, ok := value.(string); ok {
			(*flattened)[key] = value.(string)
		} else if _, ok := value.(float64); ok {
			(*flattened)[key] = value.(float64)
		} else if _, ok := value.(int); ok {
			(*flattened)[key] = value.(int)
		} else if _, ok := value.(int64); ok {
			(*flattened)[key] = value.(int64)
		} else if _, ok := value.(bool); ok {
			(*flattened)[key] = value.(bool)
		} else if _, ok := value.(time.Time); ok {
			(*flattened)[key] = value.(time.Time).Format("2006-01-02T15:04:05Z07:00")
		} else if _, ok := value.(bson.ObjectId); ok {
			(*flattened)[key] = value.(bson.ObjectId).Hex()
		} else if _, ok := value.([]interface{}); ok {
			for i := 0; i < len(value.([]interface{})); i++ {
				if _, ok := value.([]string); ok {
					stringI := string(i)
					(*flattened)[stringI] = value.(string)
				} else if _, ok := value.([]int); ok {
					stringI := string(i)
					(*flattened)[stringI] = value.(int)
				} else {
					flatten(value.([]interface{})[i].(bson.M), key+"."+strconv.Itoa(i)+".", flattened)
				}
			}
		} else {
			if value != nil {
				flatten(value.(bson.M), key+".", flattened)
			} else {
				(*flattened)[key] = ""
			}
		}
	}
}

func main() {
	flag.Parse()
	time.Local = time.UTC

	// After cmd flag parse
	file, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create Writer
	writer := csv.NewWriter(file)

	// Connect to MongoDB
	session, err := mgo.Dial(*host)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	// Set monotonic mode
	session.SetMode(mgo.Monotonic, true)

	// select DB
	db := session.DB(*dbname)

	// select Collection
	collection := db.C(*col)

	var headers []string
	if len(keys) <= 0 {
		// Auto Detect Headerline
		var h bson.M
		err = collection.Find(nil).One(&h)
		if err != nil {
			log.Fatal(err)
		}

		flatten(h, "", &flattened)
		for key := range flattened {
			headers = append(headers, key)
		}
	} else {
		// Using given fields
		headers = keys
	}

	// Default sort the headers
	// Otherwise accessing the headers will be
	// different each time.
	sort.Strings(headers)
	// write headers to file
	writer.Write(headers)
	writer.Flush()
	// log.Print(headers)

	// Create a cursor using Find query
	cursor := collection.Find(nil).Iter()
	if err != nil {
		log.Fatal(err)
	}

	// Iterate over all items in a collection
	var m bson.M
	count := 0
	for cursor.Next(&m) {
		var record []string

		flatten(m, "", &flattened)
		for _, header := range headers {
			record = append(record, fmt.Sprint(flattened[header]))
		}
		writer.Write(record)
		writer.Flush()
		count++
	}

	if cursor.Err() != nil {
		cursor.Close()
	}

	fmt.Printf("%d record(s) exported\n", count)
}
