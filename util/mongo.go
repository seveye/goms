package util

import (
	"context"
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

type Mongo struct {
	ConnStr string
	Client  *mongo.Client
	Name    string
	Indexes map[string][]mongo.IndexModel
}

func ConnectMongo(m *Mongo) error {
	var err error
	clientOptions := options.Client().ApplyURI(m.ConnStr)
	m.Client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return err
	}
	err = m.Client.Ping(context.TODO(), nil)
	if err != nil {
		return err
	}

	//如果连接参数有设置数据库名，则使用该数据库名
	cs, _ := connstring.ParseAndValidate(m.ConnStr)
	if cs.Database != "" {
		m.Name = cs.Database
	}

	//初始化不存在的索引
	initIndex(m)
	return nil
}

func initIndex(m *Mongo) {
	for name, indexes := range m.Indexes {
		collection := m.Client.Database(m.Name).Collection(name)
		indexNames := getIndexes(collection)
		for _, index := range indexes {
			indexName := getIndexName(index)
			if InArray(&indexNames, indexName) {
				continue
			}
			collection.Indexes().CreateOne(context.Background(), index)
			log.Println("创建索引", name, indexName)
		}

	}
}

//IndexName ...
type IndexName struct {
	Name string
}

func getIndexes(collection *mongo.Collection) []string {
	cur, err := collection.Indexes().List(context.Background())
	if err != nil {
		return []string{}
	}

	if err := cur.Err(); err != nil {
		return []string{}
	}

	defer cur.Close(context.Background())
	var names []string
	for cur.Next(context.Background()) {
		var im IndexName
		if err = cur.Decode(&im); err != nil {
			return []string{}
		}

		names = append(names, im.Name)
	}
	return names
}

func getIndexName(index mongo.IndexModel) string {
	var arr []string
	for _, d := range index.Keys.(bsonx.Doc) {
		arr = append(arr, d.Key)
		arr = append(arr, d.Value.String())
	}
	return strings.Join(arr, "_")
}
