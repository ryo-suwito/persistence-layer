package main

import (
    "log"
    "net"
    "persistence-layer/adapters"
    "persistence-layer/config"
    "persistence-layer/models"
    "persistence-layer/orm"
    "persistence-layer/services"
    "persistence-layer/utils"
    "google.golang.org/grpc"
    "reflect"
)

// RegisterableService is an interface that requires services to have a Register method.
type RegisterableService interface {
    Register(server *grpc.Server)
}

// GetAllServices returns a list of all RegisterableService implementations.
func GetAllServices(ormLayer *orm.ORM) []RegisterableService {
    return []RegisterableService{
        services.NewTagServiceServerImpl(ormLayer),
        services.NewPosttagServiceServerImpl(ormLayer),
        services.NewPostServiceServerImpl(ormLayer),
        services.NewUserServiceServerImpl(ormLayer),
        services.NewCategoryServiceServerImpl(ormLayer),
        services.NewCommentServiceServerImpl(ormLayer),
        services.NewProductServiceServerImpl(ormLayer),
    }

}

// RegisterAllServices dynamically registers all services that implement RegisterableService.
func RegisterAllServices(server *grpc.Server, ormLayer *orm.ORM) {
    for _, service := range GetAllServices(ormLayer) {
        serviceType := reflect.TypeOf(service).Elem().Name()
        service.Register(server)
        println("Registered service:", serviceType)
    }
}

func main() {
    // Initialize logger
    utils.InitLogger()

    // Load configuration
    cfg, err := config.LoadConfigFromFile("config/config.yaml")
    if err != nil {
        utils.LogError(err, map[string]interface{}{"context": "config"})
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // Initialize Adapters
    sqlAdapter := adapters.NewSQLAdapter(cfg.MySQLDSN, "mysql")
    mongoAdapter := adapters.NewMongoAdapter(cfg.MongoURI)
    redisAdapter := adapters.NewRedisAdapter(cfg.RedisURI)
    esAdapter := adapters.NewESAdapter(cfg.ElasticsearchURI)

    defer sqlAdapter.Close()
    defer mongoAdapter.Disconnect()
    defer redisAdapter.Close()
    defer esAdapter.Close()

    // ORM layer setup
    ormLayer := orm.NewORM(sqlAdapter, mongoAdapter, redisAdapter, esAdapter)
	                                                                    // Run GORM auto-migration for your models here
    db := sqlAdapter.GetDB()
    err = db.AutoMigrate(
        &models.Product{},
        &models.Comment{},
        &models.Posttag{},
        &models.User{},
        &models.Tag{},
        &models.Category{},
        &models.Post{},
    )
    if err != nil {
        log.Fatalf("Failed to auto migrate models: %v", err)
    }
    log.Println("Auto migration completed successfully.")
    // gRPC server setup
    grpcServer := grpc.NewServer()

    // Dynamically register all services with the gRPC server.
    RegisterAllServices(grpcServer, ormLayer)

    // Start listening on port 50051
    listener, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Failed to listen on port 50051: %v", err)
    }

    log.Println("gRPC server listening on port 50051")
    if err := grpcServer.Serve(listener); err != nil {
        log.Fatalf("Failed to serve gRPC: %v", err)
    }
}
