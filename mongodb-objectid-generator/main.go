package main

import (
	"fmt"
	"time"

	"github.com/tuananhlai/prototypes/mongodb-objectid-generator/objectid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	for range 10 {
		fmt.Println(objectid.NewString(), bson.NewObjectID().Hex())
		time.Sleep(time.Second)
	}
}
