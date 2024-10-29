package orm

import (
    "persistence-layer/adapters"
    "persistence-layer/utils"
    "time"
)

// ORM struct to integrate all data adapters.
type ORM struct {
    SQL           *adapters.SQLAdapter
    Mongo         *adapters.MongoAdapter
    Redis         *adapters.RedisAdapter
    Elasticsearch *adapters.ESAdapter
}

// NewORM initializes and returns a new ORM instance.
func NewORM(sql *adapters.SQLAdapter, mongo *adapters.MongoAdapter, redis *adapters.RedisAdapter, es *adapters.ESAdapter) *ORM {
    utils.InitLogger() // Initialize logging.
    return &ORM{
        SQL:           sql,
        Mongo:         mongo,
        Redis:         redis,
        Elasticsearch: es,
    }
}

// Create inserts a new record into the primary SQL database with transaction.
func (o *ORM) Create(model interface{}) error {
    tx, err := NewSQLTransaction(o.SQL)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    err = tx.Create(model)
    if err != nil {
        utils.LogError(err, map[string]interface{}{"operation": "Create", "model": model})
        return utils.HandleSQLError(err)
    }

    err = tx.Commit()
    if err != nil {
        utils.LogError(err, map[string]interface{}{"operation": "Create Commit", "model": model})
        return err
    }

    utils.LogInfo("Record created successfully", map[string]interface{}{"model": model})
    return nil
}

// Update updates an existing record in the primary SQL database with transaction.
func (o *ORM) Update(model interface{}) error {
    tx, err := NewSQLTransaction(o.SQL)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    err = tx.Update(model)
    if err != nil {
        utils.LogError(err, map[string]interface{}{"operation": "Update", "model": model})
        return utils.HandleSQLError(err)
    }

    err = tx.Commit()
    if err != nil {
        utils.LogError(err, map[string]interface{}{"operation": "Update Commit", "model": model})
        return err
    }

    utils.LogInfo("Record updated successfully", map[string]interface{}{"model": model})
    return nil
}

// Delete removes a record from the primary SQL database by ID with transaction.
func (o *ORM) Delete(id uint, model interface{}) error {
    tx, err := NewSQLTransaction(o.SQL)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    err = tx.Delete(id, model)
    if err != nil {
        utils.LogError(err, map[string]interface{}{"operation": "Delete", "id": id})
        return utils.HandleSQLError(err)
    }

    err = tx.Commit()
    if err != nil {
        utils.LogError(err, map[string]interface{}{"operation": "Delete Commit", "id": id})
        return err
    }

    utils.LogInfo("Record deleted successfully", map[string]interface{}{"id": id})
    return nil
}

// Read retrieves a record from the primary SQL database by ID.
func (o *ORM) Read(id uint, model interface{}) error {
    err := o.SQL.Read(id, model)
    if err != nil {
        utils.LogError(err, map[string]interface{}{"operation": "Read", "id": id})
        return utils.HandleSQLError(err)
    }
    utils.LogInfo("Record retrieved successfully", map[string]interface{}{"id": id, "model": model})
    return nil
}

// SearchSQL uses QueryBuilder for complex SQL queries.
func (o *ORM) SearchSQL(queryBuilder *utils.QueryBuilder, model interface{}) error {
    sqlQuery, params := queryBuilder.ToSQL()
    err := o.SQL.RawQuery(sqlQuery, params, model)
    if err != nil {
        utils.LogError(err, map[string]interface{}{"operation": "SearchSQL", "query": sqlQuery})
        return utils.HandleSQLError(err)
    }
    utils.LogInfo("SQL search executed successfully", map[string]interface{}{"query": sqlQuery, "params": params})
    return nil
}

// MongoRead retrieves a record from MongoDB using a filter.
func (o *ORM) MongoRead(collection string, filter map[string]interface{}, result interface{}) error {
    err := o.Mongo.Read(collection, filter, result)
    if err != nil {
        utils.LogError(err, map[string]interface{}{"operation": "MongoRead", "collection": collection, "filter": filter})
        return utils.HandleMongoError(err)
    }
    utils.LogInfo("MongoDB record retrieved successfully", map[string]interface{}{"collection": collection, "filter": filter})
    return nil
}

// Index indexes a document in Elasticsearch.
func (o *ORM) Index(index string, model interface{}) error {
    err := o.Elasticsearch.IndexDocument(index, model)
    if err != nil {
        utils.LogError(err, map[string]interface{}{"operation": "Index", "model": model})
        return err
    }
    utils.LogInfo("Document indexed successfully in Elasticsearch", map[string]interface{}{"model": model})
    return nil
}

// Search performs a search in Elasticsearch.
func (o *ORM) Search(index string, query map[string]interface{}, result interface{}) error {
    err := o.Elasticsearch.Search(index, query, result)
    if err != nil {
        utils.LogError(err, map[string]interface{}{"operation": "Search", "query": query})
        return err
    }
    utils.LogInfo("Elasticsearch search executed successfully", map[string]interface{}{"query": query})
    return nil
}

// SetCache sets a cache value with TTL in Redis.
func (o *ORM) SetCache(key string, value interface{}, ttl time.Duration) error {
    err := o.Redis.SetWithTTL(key, value, ttl)
    if err != nil {
        utils.LogError(err, map[string]interface{}{"operation": "SetCache", "key": key})
        return err
    }
    utils.LogInfo("Cache value set successfully", map[string]interface{}{"key": key, "ttl": ttl})
    return nil
}

// GetCache retrieves a cached value from Redis.
func (o *ORM) GetCache(key string, dest interface{}) error {
    err := o.Redis.Get(key, dest)
    if err != nil {
        utils.LogError(err, map[string]interface{}{"operation": "GetCache", "key": key})
        return err
    }
    utils.LogInfo("Cache value retrieved successfully", map[string]interface{}{"key": key})
    return nil
}

// DeleteCache deletes a cached value in Redis.
func (o *ORM) DeleteCache(key string) error {
    err := o.Redis.Delete(key)
    if err != nil {
        utils.LogError(err, map[string]interface{}{"operation": "DeleteCache", "key": key})
        return err
    }
    utils.LogInfo("Cache value deleted successfully", map[string]interface{}{"key": key})
    return nil
}
