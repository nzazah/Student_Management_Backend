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
	FindByID(ctx context.Context, id string) (*models.MongoAchievement, error)
	FindAll(ctx context.Context, filter bson.M) ([]*models.MongoAchievement, error)
	Update(ctx context.Context, id string, data *models.MongoAchievement) error
	SoftDelete(ctx context.Context, id string) error
	AddAttachment(ctx context.Context, id string, attachment models.AchievementAttachment) error
	UpdatePoints(ctx context.Context, id string, points int) error
}

type AchievementMongoRepository struct {
	collection *mongo.Collection
}

func NewAchievementMongoRepository(
	db *mongo.Database,
) IAchievementMongoRepository {
	return &AchievementMongoRepository{
		collection: db.Collection("achievements"),
	}
}

func (r *AchievementMongoRepository) Insert(
	ctx context.Context,
	data *models.MongoAchievement,
) (string, error) {

	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	data.DeletedAt = nil

	res, err := r.collection.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}

	oid := res.InsertedID.(primitive.ObjectID)
	return oid.Hex(), nil
}

func (r *AchievementMongoRepository) FindByID(ctx context.Context, id string) (*models.MongoAchievement, error) {
	var result models.MongoAchievement

	oid, err := primitive.ObjectIDFromHex(id)
	if err == nil {
	
		err = r.collection.FindOne(ctx, bson.M{
			"_id": oid,
			"$or": []bson.M{
				{"deletedAt": bson.M{"$exists": false}},
				{"deletedAt": nil},
			},
		}).Decode(&result)
		if err == nil {
			return &result, nil
		}
	}

	err = r.collection.FindOne(ctx, bson.M{
		"_id": id,
		"$or": []bson.M{
			{"deletedAt": bson.M{"$exists": false}},
			{"deletedAt": nil},
		},
	}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}


func (r *AchievementMongoRepository) FindAll(
	ctx context.Context,
	filter bson.M,
) ([]*models.MongoAchievement, error) {

	if filter == nil {
		filter = bson.M{}
	}

	filter["$or"] = []bson.M{
		{"deletedAt": bson.M{"$exists": false}},
		{"deletedAt": nil},
	}

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

func (r *AchievementMongoRepository) Update(
	ctx context.Context,
	id string,
	data *models.MongoAchievement,
) error {

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{
			"title":       data.Title,
			"description": data.Description,
			"details":     data.Details,
			"updatedAt":   time.Now(),
		}},
	)

	return err
}

func (r *AchievementMongoRepository) SoftDelete(
	ctx context.Context,
	id string,
) error {

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = r.collection.UpdateOne(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{
			"deletedAt": now,
			"updatedAt": now,
		}},
	)

	return err
}

func (r *AchievementMongoRepository) AddAttachment(
	ctx context.Context,
	id string,
	attachment models.AchievementAttachment,
) error {

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	attachment.UploadedAt = time.Now()

	_, err = r.collection.UpdateOne(ctx,
		bson.M{"_id": oid},
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

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{
			"points":    points,
			"updatedAt": time.Now(),
		}},
	)

	return err
}
