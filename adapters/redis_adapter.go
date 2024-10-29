package adapters

import (
    "context"
    "encoding/json"
    "time"

    "github.com/go-redis/redis/v8"
)

type RedisAdapter struct {
    client *redis.Client
    ctx    context.Context
}

// NewRedisAdapter creates a new instance of RedisAdapter.
func NewRedisAdapter(uri string) *RedisAdapter {
    opt, err := redis.ParseURL(uri)
    if err != nil {
        panic("Failed to parse Redis URI")
    }

    client := redis.NewClient(opt)
    return &RedisAdapter{
        client: client,
        ctx:    context.Background(),
    }
}

// SetWithTTL sets a key-value pair in Redis with a specified TTL (Time-To-Live).
func (r *RedisAdapter) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
    jsonData, err := json.Marshal(value)
    if err != nil {
        return err
    }

    return r.client.Set(r.ctx, key, jsonData, ttl).Err()
}

// Get retrieves a value from Redis and unmarshals it into the specified interface.
func (r *RedisAdapter) Get(key string, dest interface{}) error {
    val, err := r.client.Get(r.ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            return nil // Key does not exist, return nil to indicate a cache miss
        }
        return err
    }

    return json.Unmarshal([]byte(val), dest)
}

// Delete removes a key from Redis.
func (r *RedisAdapter) Delete(key string) error {
    return r.client.Del(r.ctx, key).Err()
}

// Exists checks if a key exists in Redis.
func (r *RedisAdapter) Exists(key string) (bool, error) {
    count, err := r.client.Exists(r.ctx, key).Result()
    if err != nil {
        return false, err
    }
    return count > 0, nil
}

// Close gracefully closes the Redis client connection.
func (r *RedisAdapter) Close() error {
    return r.client.Close()
}

// Explanation of Methods

//     SetWithTTL:
//         Takes a key, a value, and a TTL duration.
//         Marshals the value into JSON before storing it in Redis.
//         Uses the provided TTL to set an expiration time on the key.
//         Useful for caching data that should expire after a certain period.

//     Get:
//         Retrieves a JSON-encoded value from Redis using the specified key.
//         Unmarshals the retrieved JSON into the provided destination interface.
//         Returns nil for cache misses (when the key does not exist), allowing the caller to handle cache population.

//     Delete:
//         Deletes a key-value pair from Redis.
//         Useful for cache invalidation when the underlying data changes.

//     Exists:
//         Checks whether a given key exists in Redis.
//         Returns a boolean indicating the presence of the key.
//         Useful for determining if a cached version of the data already exists before performing costly operations.

//     Close:
//         Closes the Redis client connection gracefully.
//         Ensures that resources are released properly when the application is shutting down.

// Example Usage in services/user_service.go

// Update the UserService to include methods for caching user data in Redis. This allows us to cache user retrievals and invalidate the cache when user data is updated or deleted.

// go

// func (s *UserService) GetUserByID(id uint) (*models.User, error) {
//     cacheKey := fmt.Sprintf("user:%d", id)
//     var user models.User

//     // Attempt to retrieve user from cache
//     err := s.orm.Redis.Get(cacheKey, &user)
//     if err == nil {
//         return &user, nil
//     }

//     // If user is not found in cache, fetch from SQL database
//     err = s.orm.Read(id, &user)
//     if err != nil {
//         return nil, err
//     }

//     // Cache the user data with a TTL of 10 minutes
//     _ = s.orm.Redis.SetWithTTL(cacheKey, &user, 10*time.Minute)
//     return &user, nil
// }

// func (s *UserService) UpdateUser(id uint, updatedUser models.User) error {
//     user, err := s.GetUserByID(id)
//     if err != nil {
//         return err
//     }
//     user.Name = updatedUser.Name
//     user.Email = updatedUser.Email

//     // Update in SQL database
//     err = s.orm.Update(user)
//     if err != nil {
//         return err
//     }

//     // Invalidate the cached user data
//     cacheKey := fmt.Sprintf("user:%d", id)
//     _ = s.orm.Redis.Delete(cacheKey)

//     // Optionally, re-cache the updated user
//     _ = s.orm.Redis.SetWithTTL(cacheKey, user, 10*time.Minute)

//     return nil
// }

// func (s *UserService) DeleteUser(id uint) error {
//     var user models.User
//     err := s.orm.Delete(id, &user)
//     if err != nil {
//         return err
//     }

//     // Invalidate the cached user data
//     cacheKey := fmt.Sprintf("user:%d", id)
//     _ = s.orm.Redis.Delete(cacheKey)

//     return nil
// }

// Explanation of Cache Usage in UserService

//     GetUserByID:
//         Checks Redis for cached user data using a cache key like user:123.
//         If the user data is found in Redis, it is returned directly, avoiding a database query.
//         If not found in the cache, it retrieves the user from the SQL database and caches the result with a 10-minute TTL.

//     UpdateUser:
//         Updates the user in the SQL database.
//         Deletes the corresponding cache entry to ensure the cache is invalidated.
//         Optionally, it re-caches the updated user data.

//     DeleteUser:
//         Deletes the user from the SQL database.
//         Removes the cached user data from Redis to prevent stale cache entries.

