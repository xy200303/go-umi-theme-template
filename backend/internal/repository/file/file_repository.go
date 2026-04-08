package filerepo

import (
	"backend/internal/models/entities"
	sharedrepo "backend/internal/repository/shared"
	"strings"
	"time"

	"gorm.io/gorm"
)

type AdminFileFilter struct {
	Keyword      string
	UploadStatus string
}

type AdminFileStats struct {
	TotalCount    int64
	UploadedCount int64
	BoundCount    int64
	DeletedCount  int64
}

type FileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) FindByID(fileID string) (*entities.File, error) {
	var item entities.File
	if err := r.db.First(&item, "id = ?", fileID).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *FileRepository) Create(item *entities.File) error {
	return r.db.Create(item).Error
}

func (r *FileRepository) CreateTx(tx *gorm.DB, item *entities.File) error {
	return tx.Create(item).Error
}

func (r *FileRepository) Update(item *entities.File) error {
	return r.db.Save(item).Error
}

func (r *FileRepository) UpdateTx(tx *gorm.DB, item *entities.File) error {
	return tx.Save(item).Error
}

func (r *FileRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *FileRepository) ListStaleUploadedBefore(before time.Time, limit int) ([]entities.File, error) {
	var items []entities.File
	query := r.db.
		Where("upload_status = ?", "uploaded").
		Where("created_at < ?", before).
		Order("created_at asc")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *FileRepository) ListAdminPage(filter AdminFileFilter, page, pageSize int) ([]entities.File, int64, int, int, error) {
	page, pageSize = sharedrepo.NormalizePageParams(page, pageSize)

	query := r.applyAdminFilter(r.db.Model(&entities.File{}), filter)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, page, pageSize, err
	}

	var items []entities.File
	err := query.
		Order("created_at desc").
		Order("id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error

	return items, total, page, pageSize, err
}

func (r *FileRepository) GetAdminStats(filter AdminFileFilter) (*AdminFileStats, error) {
	query := r.applyAdminFilter(r.db.Model(&entities.File{}), filter)

	stats := &AdminFileStats{}
	if err := query.Count(&stats.TotalCount).Error; err != nil {
		return nil, err
	}

	type statusCountRow struct {
		UploadStatus string
		Count        int64
	}

	var rows []statusCountRow
	if err := query.
		Select("upload_status, COUNT(*) AS count").
		Group("upload_status").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		switch strings.TrimSpace(row.UploadStatus) {
		case "uploaded":
			stats.UploadedCount = row.Count
		case "bound":
			stats.BoundCount = row.Count
		case "deleted":
			stats.DeletedCount = row.Count
		}
	}

	return stats, nil
}

func (r *FileRepository) applyAdminFilter(query *gorm.DB, filter AdminFileFilter) *gorm.DB {
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("id ILIKE ? OR original_name ILIKE ? OR storage_path ILIKE ?", like, like, like)
	}
	if uploadStatus := strings.TrimSpace(filter.UploadStatus); uploadStatus != "" {
		query = query.Where("upload_status = ?", uploadStatus)
	}
	return query
}
