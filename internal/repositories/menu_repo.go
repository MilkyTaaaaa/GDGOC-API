package repositories

import (
	"fmt"
	"math"
	"GDGOC-API/internal/models"
	"strings"

	"gorm.io/gorm"
)

// ngehandle semua operasi database untuk menus
type MenuRepository struct {
	db *gorm.DB
}

// create instance baru
func NewMenuRepository(db *gorm.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

// insert a menu baru ke db
func (r *MenuRepository) Create(menu *models.Menu) error {
	return r.db.Create(menu).Error
}

// return semua menu dgn opsi filters & pagination
func (r *MenuRepository) GetAll(filters models.MenuFilters) ([]models.Menu, *models.PaginationMeta, error) {
	var menus []models.Menu
	var total int64

	query := r.db.Model(&models.Menu{})

	query = r.applyFilters(query, filters)

	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	query = r.applySorting(query, filters.Sort)

	page := filters.Page
	perPage := filters.PerPage

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10 // default
	}
	if perPage > 100 {
		perPage = 100 // max limit
	}

	offset := (page - 1) * perPage
	query = query.Offset(offset).Limit(perPage)

	if err := query.Find(&menus).Error; err != nil {
		return nil, nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	pagination := &models.PaginationMeta{
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}

	return menus, pagination, nil
}

// GET berdasar ID
func (r *MenuRepository) GetByID(id uint) (*models.Menu, error) {
	var menu models.Menu
	err := r.db.First(&menu, id).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

// update menu yg sudah ada
func (r *MenuRepository) Update(id uint, menu *models.Menu) error {
	var existing models.Menu
	if err := r.db.First(&existing, id).Error; err != nil {
		return err
	}
	menu.ID = id
	return r.db.Save(menu).Error
}

//  hapus menu
func (r *MenuRepository) Delete(id uint) error {
	var menu models.Menu
	if err := r.db.First(&menu, id).Error; err != nil {
		return err
	}

	return r.db.Delete(&menu).Error
}

// ngelompokin menu by category
func (r *MenuRepository) GroupByCategory(mode string, perCategory int) (interface{}, error) {
	if mode == "count" {
		type CategoryCount struct {
			Category string `json:"category"`
			Count    int64  `json:"count"`
		}

		var results []CategoryCount
		err := r.db.Model(&models.Menu{}).
			Select("category, COUNT(*) as count").
			Group("category").
			Scan(&results).Error

		if err != nil {
			return nil, err
		}

		countMap := make(map[string]int64)
		for _, result := range results {
			countMap[result.Category] = result.Count
		}

		return countMap, nil
	}

	// Return list menu per kategori
	var menus []models.Menu
	query := r.db.Model(&models.Menu{})

	if perCategory > 0 {
		var categories []string
		r.db.Model(&models.Menu{}).
			Distinct("category").
			Pluck("category", &categories)

		var allMenus []models.Menu
		for _, category := range categories {
			var categoryMenus []models.Menu
			r.db.Where("category = ?", category).
				Limit(perCategory).
				Find(&categoryMenus)
			allMenus = append(allMenus, categoryMenus...)
		}
		menus = allMenus
	} else {
		// Get semua menu
		query.Find(&menus)
	}

	grouped := make(map[string][]models.Menu)
	for _, menu := range menus {
		grouped[menu.Category] = append(grouped[menu.Category], menu)
	}

	return grouped, nil
}

// Search
func (r *MenuRepository) Search(query string, page, perPage int) ([]models.Menu, *models.PaginationMeta, error) {
	var menus []models.Menu
	var total int64

	searchQuery := r.db.Model(&models.Menu{})

	if query != "" {
		// cari by nama, deskripsi, bahan
		searchPattern := "%" + strings.ToLower(query) + "%"
		searchQuery = searchQuery.Where(
			"LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR EXISTS (SELECT 1 FROM unnest(ingredients) AS ing WHERE LOWER(ing) LIKE ?)",
			searchPattern, searchPattern, searchPattern,
		)
	}

	// hitung total
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// Set defaults for pagination
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	offset := (page - 1) * perPage
	searchQuery = searchQuery.Offset(offset).Limit(perPage)

	// eksekusi query
	if err := searchQuery.Find(&menus).Error; err != nil {
		return nil, nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	pagination := &models.PaginationMeta{
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}

	return menus, pagination, nil
}

// filter query
func (r *MenuRepository) applyFilters(query *gorm.DB, filters models.MenuFilters) *gorm.DB {
	if filters.Query != "" {
		searchPattern := "%" + strings.ToLower(filters.Query) + "%"
		query = query.Where(
			"LOWER(name) LIKE ? OR LOWER(description) LIKE ?",
			searchPattern, searchPattern,
		)
	}

	// filter kategori
	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}

	// filter harga
	if filters.MinPrice > 0 {
		query = query.Where("price >= ?", filters.MinPrice)
	}
	if filters.MaxPrice > 0 {
		query = query.Where("price <= ?", filters.MaxPrice)
	}

	// filter kalori
	if filters.MaxCalories > 0 {
		query = query.Where("calories <= ?", filters.MaxCalories)
	}

	return query
}

// sorting query
func (r *MenuRepository) applySorting(query *gorm.DB, sort string) *gorm.DB {
	if sort == "" {
		// default sorting
		return query.Order("created_at DESC")
	}

	parts := strings.Split(sort, ":")
	if len(parts) != 2 {
		return query.Order("created_at DESC")
	}

	field := parts[0]
	direction := strings.ToUpper(parts[1])

	if direction != "ASC" && direction != "DESC" {
		direction = "ASC"
	}

	validFields := map[string]bool{
		"name":       true,
		"category":   true,
		"price":      true,
		"calories":   true,
		"created_at": true,
	}

	if validFields[field] {
		return query.Order(fmt.Sprintf("%s %s", field, direction))
	}

	return query.Order("created_at DESC")
}