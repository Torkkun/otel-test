package repository

import (
	"context"
	"otel-test/database"
	"otel-test/server/entity"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type UserRepository struct {
	db     *database.DB
	tracer trace.Tracer
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{
		db:     db,
		tracer: otel.Tracer("user-repository"),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	// カスタムスパンを作成（詳細な追跡のため）
	ctx, span := r.tracer.Start(ctx, "UserRepository.Create")
	defer span.End()

	// スパンに属性を追加
	span.SetAttributes(
		attribute.String("operation", "create_user"),
		attribute.String("user.email", user.Email),
	)

	// GORMでコンテキストを使用（自動的にトレースされる）
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.SetAttributes(attribute.Int("user.id", int(user.ID)))
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.GetByID")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation", "get_user_by_id"),
		attribute.Int("user.id", int(id)),
	)

	var user entity.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.String("user.email", user.Email))
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.GetByEmail")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation", "get_user_by_email"),
		attribute.String("user.email", email),
	)

	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]entity.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.List")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation", "list_users"),
		attribute.Int("query.limit", limit),
		attribute.Int("query.offset", offset),
	)

	var users []entity.User
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("result.count", len(users)))
	return users, nil
}
