package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// users table
	User struct {
		ID    uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
		Name  string    `gorm:"column:name;type:varchar(255);not null"`
		Email string    `gorm:"column:email;type:varchar(255);not null"`
	}

	CreateUserParams struct {
		Name  string
		Email string
	}
)

func (User) TableName() string {
	return "users"
}

func (r *Repository) GetUsers(ctx context.Context) ([]*User, error) {
	users := []*User{}
	if err := r.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("select users: %w", err)
	}

	return users, nil
}

func (r *Repository) CreateUser(ctx context.Context, params CreateUserParams) (uuid.UUID, error) {
	user := User{
		ID:    uuid.New(),
		Name:  params.Name,
		Email: params.Email,
	}

	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		return uuid.Nil, fmt.Errorf("insert user: %w", err)
	}

	return user.ID, nil
}

func (r *Repository) GetUser(ctx context.Context, userID uuid.UUID) (*User, error) {
	user := new(User)
	if err := r.db.WithContext(ctx).First(user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("select user: %w", err)
		}

		return nil, fmt.Errorf("select user: %w", err)
	}

	return user, nil
}
