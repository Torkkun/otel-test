package service

import (
	"context"
	"fmt"
	"otel-test/server/entity"
	"otel-test/server/repository"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo *repository.UserRepository
	tracer   trace.Tracer
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		tracer:   otel.Tracer("user-service"),
	}
}

func (s *UserService) CreateUser(ctx context.Context, name, email string) (*entity.User, error) {
	ctx, span := s.tracer.Start(ctx, "UserService.CreateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.name", name),
		attribute.String("user.email", email),
	)

	// 既存ユーザーチェック
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		span.SetAttributes(attribute.Bool("user.already_exists", true))
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	// 新規ユーザー作成
	user := &entity.User{
		Name:  name,
		Email: email,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(attribute.Int("user.created_id", int(user.ID)))
	return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id uint) (*entity.User, error) {
	ctx, span := s.tracer.Start(ctx, "UserService.GetUserByID")
	defer span.End()

	span.SetAttributes(attribute.Int("user.id", int(id)))

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			span.SetAttributes(attribute.Bool("user.not_found", true))
			return nil, nil // ユーザーが見つからない場合はnilを返す
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]entity.User, error) {
	ctx, span := s.tracer.Start(ctx, "UserService.ListUsers")
	defer span.End()

	span.SetAttributes(
		attribute.Int("query.limit", limit),
		attribute.Int("query.offset", offset),
	)

	users, err := s.userRepo.List(ctx, limit, offset)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	span.SetAttributes(attribute.Int("result.count", len(users)))
	return users, nil
}
