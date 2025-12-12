package repositories

import (
    "context"
    "time"
    "uas/app/models"

    "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

    "go.mongodb.org/mongo-driver/mongo"
)

type IAchievementMongoRepository interface {
    Insert(ctx context.Context, data *models.MongoAchievement) (string, error)
    SoftDelete(ctx context.Context, id string) error
}

type AchievementMongoRepository struct {
    Col *mongo.Collection
}

func NewAchievementMongoRepository(client *mongo.Client) IAchievementMongoRepository {
    return &AchievementMongoRepository{
        Col: client.Database("uas").Collection("achievements"),
    }
}

func (r *AchievementMongoRepository) Insert(ctx context.Context, data *models.MongoAchievement) (string, error) {

    // generate ID manual (string) biar aman di Postgres juga
    oid := primitive.NewObjectID()
    data.ID = oid.Hex()

    data.CreatedAt = time.Now()
    data.UpdatedAt = time.Now()

    // insert ke Mongo
    _, err := r.Col.InsertOne(ctx, data)
    if err != nil {
        return "", err
    }

    return data.ID, nil
}

func (r *AchievementMongoRepository) SoftDelete(ctx context.Context, id string) error {
    now := time.Now()

    _, err := r.Col.UpdateOne(
        ctx,
        bson.M{"_id": id},
        bson.M{"$set": bson.M{"deletedAt": now}},
    )

    return err
}
