package adapters

import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "time"
)

type MongoAdapter struct {
    client *mongo.Client
    ctx    context.Context
}

// NewMongoAdapter initializes a new MongoAdapter with a given URI.
func NewMongoAdapter(uri string) *MongoAdapter {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        panic("Failed to connect to MongoDB")
    }
    return &MongoAdapter{client: client, ctx: context.TODO()}
}

// Create inserts a new document into a MongoDB collection.
func (m *MongoAdapter) Create(collection string, model interface{}) error {
    col := m.client.Database("app_db").Collection(collection)
    _, err := col.InsertOne(m.ctx, model)
    return err
}

// Read retrieves a document from a MongoDB collection using a filter.
func (m *MongoAdapter) Read(collection string, filter map[string]interface{}, result interface{}) error {
    col := m.client.Database("app_db").Collection(collection)
    return col.FindOne(m.ctx, filter).Decode(result)
}

// Update modifies an existing document in a MongoDB collection using a filter.
func (m *MongoAdapter) Update(collection string, filter map[string]interface{}, update interface{}) error {
    col := m.client.Database("app_db").Collection(collection)
    _, err := col.UpdateOne(m.ctx, filter, bson.M{"$set": update})
    return err
}

// Delete removes a document from a MongoDB collection using a filter.
func (m *MongoAdapter) Delete(collection string, filter map[string]interface{}) error {
    col := m.client.Database("app_db").Collection(collection)
    _, err := col.DeleteOne(m.ctx, filter)
    return err
}

// Disconnect closes the MongoDB connection.
func (m *MongoAdapter) Disconnect() {
    _ = m.client.Disconnect(m.ctx)
}
