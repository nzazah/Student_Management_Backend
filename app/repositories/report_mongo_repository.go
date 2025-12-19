package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type IAchievementMongoReportRepository interface {
	SumPointsByIDs(ctx context.Context, ids []string) (int, error)
	CountByType(ctx context.Context, ids []string) (map[string]int, error)
	FindByIDs(ctx context.Context, ids []string) ([]bson.M, error)
}

type achievementMongoReportRepository struct {
	collection *mongo.Collection
}

func NewAchievementMongoReportRepository(db *mongo.Database) IAchievementMongoReportRepository {
	return &achievementMongoReportRepository{
		collection: db.Collection("achievement_collections"),
	}
}

func (r *achievementMongoReportRepository) SumPointsByIDs(ctx context.Context, ids []string) (int, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": ids}}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$points"}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result []struct {
		Total int `bson:"total"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
		return result[0].Total, nil
	}

	return 0, nil
}

func (r *achievementMongoReportRepository) CountByType(ctx context.Context, ids []string) (map[string]int, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": ids}}},
		{"$group": bson.M{"_id": "$achievementType", "count": bson.M{"$sum": 1}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]int)
	for cursor.Next(ctx) {
		var row struct {
			Type  string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&row); err != nil {
			return nil, err
		}
		result[row.Type] = row.Count
	}

	return result, nil
}

func (r *achievementMongoReportRepository) FindByIDs(ctx context.Context, ids []string) ([]bson.M, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
