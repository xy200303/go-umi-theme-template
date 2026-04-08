package systemrepo

import (
	"strings"

	"backend/internal/models/entities"

	"gorm.io/gorm"
)

type AuditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(item *entities.AuditLog) error {
	return r.db.Create(item).Error
}

func (r *AuditLogRepository) List(keyword string, menuKey string, statusCode int, page int, pageSize int) ([]entities.AuditLog, int64, error) {
	var (
		list  []entities.AuditLog
		total int64
	)

	query := r.db.Model(&entities.AuditLog{})
	if trimmedMenuKey := strings.TrimSpace(menuKey); trimmedMenuKey != "" {
		query = query.Where("menu_key = ?", trimmedMenuKey)
	}
	if statusCode > 0 {
		query = query.Where("status_code = ?", statusCode)
	}
	if trimmed := strings.TrimSpace(keyword); trimmed != "" {
		like := "%" + trimmed + "%"
		query = query.Where(
			"username ILIKE ? OR operation_id ILIKE ? OR operation_name ILIKE ? OR method ILIKE ? OR route_path ILIKE ? OR request_path ILIKE ? OR client_ip ILIKE ?",
			like,
			like,
			like,
			like,
			like,
			like,
			like,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at desc").Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *AuditLogRepository) PruneExcess(limit int) error {
	if limit <= 0 {
		return nil
	}

	subQuery := r.db.Model(&entities.AuditLog{}).
		Select("id").
		Order("created_at desc").
		Order("id desc").
		Offset(limit)

	return r.db.Where("id IN (?)", subQuery).Delete(&entities.AuditLog{}).Error
}
