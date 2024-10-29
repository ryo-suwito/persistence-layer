package utils

import (
    "errors"
    "go.mongodb.org/mongo-driver/mongo"
    "gorm.io/gorm"
)

var (
    ErrNotFound = errors.New("record not found")
    ErrDatabase = errors.New("database error")
)

func HandleSQLError(err error) error {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return ErrNotFound
    }
    return ErrDatabase
}

func HandleMongoError(err error) error {
    if errors.Is(err, mongo.ErrNoDocuments) {
        return ErrNotFound
    }
    return ErrDatabase
}
