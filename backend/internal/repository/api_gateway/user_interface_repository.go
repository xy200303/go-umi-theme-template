package apigatewayrepo

import (
	"backend/internal/models/entities"
	"time"

	"gorm.io/gorm"
)

type UserInterfaceRepository struct {
	db *gorm.DB
}

func NewUserInterfaceRepository(db *gorm.DB) *UserInterfaceRepository {
	return &UserInterfaceRepository{db: db}
}

func (r *UserInterfaceRepository) ListByUserID(userID uint) ([]entities.UserInterface, error) {
	var items []entities.UserInterface
	err := r.db.Where("user_id = ?", userID).Order("id desc").Find(&items).Error
	return items, err
}

func (r *UserInterfaceRepository) FindByIDForUser(id uint, userID uint) (*entities.UserInterface, error) {
	var item entities.UserInterface
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *UserInterfaceRepository) FindByGatewayKeyHash(hash string) (*entities.UserInterface, error) {
	var item entities.UserInterface
	err := r.db.Where("gateway_key_hash = ?", hash).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *UserInterfaceRepository) Create(item *entities.UserInterface) error {
	return r.db.Create(item).Error
}

func (r *UserInterfaceRepository) Update(item *entities.UserInterface) error {
	return r.db.Save(item).Error
}

func (r *UserInterfaceRepository) Delete(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entities.UserInterface{}).Error
}

func (r *UserInterfaceRepository) TouchLastUsedAt(id uint, lastUsedAt time.Time) error {
	return r.db.Model(&entities.UserInterface{}).
		Where("id = ?", id).
		Update("last_used_at", lastUsedAt).
		Error
}
