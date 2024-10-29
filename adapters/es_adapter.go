package adapters

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"

    "github.com/elastic/go-elasticsearch/v8"
    "github.com/elastic/go-elasticsearch/v8/esapi"
)

type ESAdapter struct {
    client *elasticsearch.Client
    ctx    context.Context
}

// NewESAdapter initializes a new Elasticsearch adapter with a given URI.
func NewESAdapter(uri string) *ESAdapter {
    cfg := elasticsearch.Config{Addresses: []string{uri}}
    client, err := elasticsearch.NewClient(cfg)
    if err != nil {
        panic("Failed to connect to Elasticsearch")
    }
    return &ESAdapter{
        client: client,
        ctx:    context.Background(),
    }
}

// IndexDocument indexes a model into Elasticsearch.
func (e *ESAdapter) IndexDocument(index string, model interface{}) error {
    body, err := json.Marshal(model)
    if err != nil {
        return err
    }

    // Use reflection to extract the ID field from the model.
    id, err := extractID(model)
    if err != nil {
        return err
    }

    req := esapi.IndexRequest{
        Index:      index,
        DocumentID: id,
        Body:       bytes.NewReader(body),
        Refresh:    "true",
    }

    res, err := req.Do(e.ctx, e.client)
    if err != nil {
        return err
    }
    defer res.Body.Close()

    if res.IsError() {
        return errors.New("error indexing document: " + res.String())
    }

    return nil
}

// UpdateDocument updates an existing document in Elasticsearch.
func (e *ESAdapter) UpdateDocument(index string, model interface{}) error {
    body, err := json.Marshal(map[string]interface{}{"doc": model})
    if err != nil {
        return err
    }

    id, err := extractID(model)
    if err != nil {
        return err
    }

    req := esapi.UpdateRequest{
        Index:      index,
        DocumentID: id,
        Body:       bytes.NewReader(body),
        Refresh:    "true",
    }

    res, err := req.Do(e.ctx, e.client)
    if err != nil {
        return err
    }
    defer res.Body.Close()

    if res.IsError() {
        return errors.New("error updating document: " + res.String())
    }

    return nil
}

// Search executes a search query in Elasticsearch.
func (e *ESAdapter) Search(index string, query map[string]interface{}, result interface{}) error {
    body, err := json.Marshal(query)
    if err != nil {
        return err
    }

    res, err := e.client.Search(
        e.client.Search.WithContext(e.ctx),
        e.client.Search.WithIndex(index),
        e.client.Search.WithBody(bytes.NewReader(body)),
        e.client.Search.WithTrackTotalHits(true),
    )
    if err != nil {
        return err
    }
    defer res.Body.Close()

    if res.IsError() {
        return errors.New("error executing search: " + res.String())
    }

    return json.NewDecoder(res.Body).Decode(result)
}

// DeleteDocument removes a document from Elasticsearch.
func (e *ESAdapter) DeleteDocument(index string, model interface{}) error {
    id, err := extractID(model)
    if err != nil {
        return err
    }

    req := esapi.DeleteRequest{
        Index:      index,
        DocumentID: id,
        Refresh:    "true",
    }

    res, err := req.Do(e.ctx, e.client)
    if err != nil {
        return err
    }
    defer res.Body.Close()

    if res.IsError() {
        return errors.New("error deleting document: " + res.String())
    }

    return nil
}

// extractID extracts the ID as a string from a model struct.
func extractID(model interface{}) (string, error) {
    if m, ok := model.(interface{ GetID() uint64 }); ok {
        return string(m.GetID()), nil
    }
    return "", errors.New("model does not have a GetID method")
}

// Close is a placeholder for compatibility but doesn't need to close anything for Elasticsearch.
func (e *ESAdapter) Close() error {
    return nil
}


// Example of Query Map for Search

// A query map in Go can be used to perform a search. Here’s an example of how you might construct a query for searching by a user's name:

// query := map[string]interface{}{
//     "query": map[string]interface{}{
//         "match": map[string]interface{}{
//             "name": "John Doe",
//         },
//     },
// }

// Using the Elasticsearch Adapter in services/user_service.go

// Update the UserService to include methods that leverage the Elasticsearch adapter for full-text searches.

// go

// func (s *UserService) SearchUsersByName(name string) ([]models.User, error) {
//     query := map[string]interface{}{
//         "query": map[string]interface{}{
//             "match": map[string]interface{}{
//                 "name": name,
//             },
//         },
//     }

//     users, err := s.orm.Search(query)
//     if err != nil {
//         return nil, err
//     }

//     return users, nil
// }

// func (s *UserService) IndexUser(user models.User) error {
//     return s.orm.Index(user)
// }

// func (s *UserService) DeleteUserIndex(id uint) error {
//     return s.orm.Delete(id)
// }