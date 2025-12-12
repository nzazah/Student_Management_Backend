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
    FindByID(ctx context.Context, id string) (*models.MongoAchievement, error)
	FindAll(ctx context.Context, filter bson.M) ([]*models.MongoAchievement, error)
	Update(ctx context.Context, id string, data *models.MongoAchievement) error
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

func (r *AchievementMongoRepository) FindByID(ctx context.Context, id string) (*models.MongoAchievement, error) {
	var result models.MongoAchievement
	err := r.Col.FindOne(ctx, bson.M{"_id": id, "deletedAt": nil}).Decode(&result)
	return &result, err
}

func (r *AchievementMongoRepository) FindAll(ctx context.Context, filter bson.M) ([]*models.MongoAchievement, error) {
	filter["deletedAt"] = nil
	cursor, err := r.Col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*models.MongoAchievement
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *AchievementMongoRepository) Update(ctx context.Context, id string, data *models.MongoAchievement) error {
	data.UpdatedAt = time.Now()
	_, err := r.Col.UpdateOne(ctx,
		bson.M{"_id": id, "deletedAt": nil},
		bson.M{"$set": bson.M{
			"title":       data.Title,
			"description": data.Description,
			"details":     data.Details,
			"attachments": data.Attachments,
			"updatedAt":   data.UpdatedAt,
		}},
	)
	return err
}
