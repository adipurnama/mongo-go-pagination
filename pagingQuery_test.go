package mongopagination_test

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	pagination "github.com/adipurnama/mongo-go-pagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TodoTest struct {
	Title     string    `json:"title" bson:"title"`
	Status    string    `json:"status" bson:"status"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

const (
	DatabaseHost       string = "mongodb://root:example@localhost:27017"
	DatabaseName       string = "todo"
	DatabaseCollection string = "TodoTest"
)

func cleanup(db *mongo.Database) (err error) {
	err = db.Collection(DatabaseCollection).Drop(context.Background())
	return
}

func insertExamples(db *mongo.Database) (insertedIds []interface{}, err error) {
	var data []interface{}
	for i := 0; i < 20; i++ {
		data = append(data, bson.M{
			"title":     fmt.Sprintf("todo-%d", i),
			"status":    "active",
			"createdAt": time.Now(),
		})
	}
	result, err := db.Collection(DatabaseCollection).InsertMany(
		context.Background(), data)
	if err != nil {
		return nil, err
	}
	return result.InsertedIDs, nil
}

func TestPagingQuery_Find(t *testing.T) {
	_, session := NewConnection()
	db := session.Database(DatabaseName)
	defer cleanup(db)
	insertedIds, err := insertExamples(db)
	if len(insertedIds) < 1 {
		t.Errorf("Empty insert ids")
	}
	if err != nil {
		t.Errorf("Data insert error. Error: %s", err.Error())
	}
	filter := bson.M{}

	var limit int64 = 10
	var page int64

	projection := bson.D{
		{"title", 1},
		{"status", 1},
	}
	collection := db.Collection(DatabaseCollection)
	paginatedData, err := pagination.New(collection).Limit(limit).Page(page).Sort("price", -1).Select(projection).Filter(filter).Find(context.TODO())

	if err != nil {
		t.Errorf("Error while pagination. Error: %s", err.Error())
	}
	if paginatedData == nil {
		t.Errorf("Empty Pagination data error")
		return
	}

	if len(paginatedData.Data) < 1 {
		t.Errorf("Error fetching data")
	}

	if paginatedData.Pagination.Total != 20 || paginatedData.Pagination.Page != 1 {
		t.Errorf("False Pagination data should be 20 but got: %d", paginatedData.Pagination.Total)
	}

	// no limit or page provided error
	_, noLimitOrPageError := pagination.New(collection).Sort("price", -1).Select(projection).Filter(filter).Find(context.TODO())
	if noLimitOrPageError == nil {
		t.Errorf("Error expected but got no error")
	}

	// no filter error
	_, noFilterError := pagination.New(collection).Limit(limit).Page(page).Sort("price", -1).Select(projection).Find(context.TODO())
	if noFilterError == nil {
		t.Errorf("Error expected but got no error")
	}

	// getting page 2 data
	page = 2
	limit = 0 // defaults to 10

	// Aggregate pipeline pagination test
	match := bson.M{"$match": bson.M{"status": "active"}}

	aggPaginatedData, err := pagination.New(collection).Limit(limit).Page(page).Sort("price", -1).Aggregate(context.TODO(), match)
	if err != nil {
		t.Errorf("Error while Aggregation pagination. Error: %s", err.Error())
	}

	if aggPaginatedData == nil {
		t.Errorf("Empty Aggregated Pagination data error")
		return
	}

	// Aggregation error match query test
	faultyMatch := bson.M{"$matches": bson.M{"status": "active"}}
	_, faultyMatchQuery := pagination.New(collection).Sort("price", -1).Aggregate(context.TODO(), faultyMatch)
	if faultyMatchQuery == nil {
		t.Errorf("Error expected but got no error")
	}

	// no limit or page provided error
	_, noLimitOrPageAggError := pagination.New(collection).Sort("price", -1).Aggregate(context.TODO(), match)
	if noLimitOrPageAggError == nil {
		t.Errorf("Error expected but got no error")
	}

	// filter in aggregate error
	_, noFilterAggError := pagination.New(collection).Limit(limit).Page(page).Filter(filter).Sort("price", -1).Aggregate(context.TODO(), match)
	if noFilterAggError == nil {
		t.Errorf("Error expected but got no error")
	}

	// without sorting test
	_, sortProvideTest := pagination.New(collection).Aggregate(context.TODO(), match)
	if sortProvideTest == nil {
		t.Errorf("data expected")
		return
	}
}

func NewConnection() (a *mongo.Database, b *mongo.Client) {
	var connectOnce sync.Once
	var db *mongo.Database
	var session *mongo.Client
	connectOnce.Do(func() {
		db, session = connect()
	})

	return db, session
}

func connect() (a *mongo.Database, b *mongo.Client) {
	var err error
	session, err := mongo.NewClient(options.Client().ApplyURI(DatabaseHost))
	if err != nil {
		log.Fatal(err)
	}
	err = session.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	var db = session.Database(DatabaseName)
	return db, session
}
