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
    AddAttachment(ctx context.Context, id string, attachment models.AchievementAttachment) error
	UpdatePoints(ctx context.Context, id string, points int) error
}


type AchievementMongoRepository struct {
	collection *mongo.Collection
}


func NewAchievementMongoRepository(client *mongo.Client) IAchievementMongoRepository {
	return &AchievementMongoRepository{
		collection: client.Database("uas").Collection("achievements"),
	}
}


func (r *AchievementMongoRepository) Insert(ctx context.Context, data *models.MongoAchievement) (string, error) {
	oid := primitive.NewObjectID()
	data.ID = oid.Hex()
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}

	return data.ID, nil
}


func (r *AchievementMongoRepository) SoftDelete(ctx context.Context, id string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"deletedAt": time.Now()}},
	)
	return err
}


func (r *AchievementMongoRepository) FindByID(ctx context.Context, id string) (*models.MongoAchievement, error) {
	var result models.MongoAchievement
	err := r.collection.FindOne(
		ctx,
		bson.M{"_id": id, "deletedAt": nil},
	).Decode(&result)
	return &result, err
}


func (r *AchievementMongoRepository) FindAll(ctx context.Context, filter bson.M) ([]*models.MongoAchievement, error) {
	filter["deletedAt"] = nil

	cursor, err := r.collection.Find(ctx, filter)
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

	_, err := r.collection.UpdateOne(
		ctx,
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

func (r *AchievementMongoRepository) AddAttachment(
	ctx context.Context,
	id string,
	attachment models.AchievementAttachment,
) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id, "deletedAt": nil},
		bson.M{
			"$push": bson.M{"attachments": attachment},
			"$set":  bson.M{"updatedAt": time.Now()},
		},
	)
	return err
}

func (r *AchievementMongoRepository) UpdatePoints(
	ctx context.Context,
	id string,
	points int,
) error {

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"points":    points,
				"updatedAt": time.Now(),
			},
		},
	)

	return err
}
