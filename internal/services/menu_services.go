package services

import(
	"errors"
	"GDGOC-API/internal/models"
	"GDGOC-API/internal/repositories"

	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type MenuService struct{
	repo	*repositories.MenuRepository
	validate	*validator.Validate
}

func NewMenuService(repo *repositories.MenuRepository) *MenuService{
	return &MenuService{
		repo:	repo,
		validate:	validator.New(),
	}
}

// create menu baru
func (s *MenuService) CreateMenu(req models.CreateMenuRequest) (*models.Menu, error){
	if err := s.validate.Struct(req); err != nil{
		return nil, err
	}

	menu := &models.Menu{
		Name:	req.Name,
		Category:	req.Category,
		Calories:	req.Calories,
		Price:	req.Price,
		Ingredients:	pq.StringArray(req.Ingredients),
		Description:	req.Description,
	}
	if err := s.repo.Create(menu); err != nil{
		return nil, err
	}
	return menu, nil
}

// get semua menu w/ filter & pagination
func (s *MenuService) GetAllMenus(filters models.MenuFilters) ([]models.Menu, *models.PaginationMeta, error){
	if filters.Page < 1{
		filters.Page = 1
	}
	if filters.PerPage < 1{
		filters.PerPage = 10
	}
	if filters.PerPage > 100{
		filters.PerPage = 100
	}

	menus, pagination, err := s.repo.GetAll(filters)
	if err != nil{
		return nil, nil, err
	}
	return menus,pagination, nil
}

// get menu by id
func (s *MenuService) GetMenuByID(id uint) (*models.Menu, error){
	menu, err := s.repo.GetByID(id)
	if err != nil{
		if errors.Is(err, gorm.ErrRecordNotFound){
			return nil, errors.New("menu tidak ditemukan")
		}
		return nil, err
	}
	return menu, nil
}

// update menu
func (s *MenuService) UpdateMenu(id uint, req models.UpdateMenuRequest) (*models.Menu, error){
	if err := s.validate.Struct(req); err != nil{
		return nil, err
	}

	// cek ketersediaan menu
	existing, err := s.repo.GetByID(id)
	if err != nil{
		if errors.Is(err, gorm.ErrRecordNotFound){
			return nil, errors.New("menu tidak ditemukan")
		}
		return nil, err
	}

	//update field
	existing.Name = req.Name
	existing.Category = req.Category
	existing.Calories= req.Calories
	existing.Price= req.Price
	existing.Ingredients = pq.StringArray(req.Ingredients)
	existing.Description= req.Description

	if err := s.repo.Update(id, existing); err != nil{
		return nil, err
	}

	return existing, nil
}

// hapus menu by id
func (s *MenuService) DeleteMenu(id uint) error{
	_, err := s.repo.GetByID(id)
	if err != nil{
		if errors.Is(err, gorm.ErrRecordNotFound){
			return errors.New("menu tidak ditemukan")
		}
	return err
	}

	return s.repo.Delete(id)
}

// grouping menu by kategori
func (s *MenuService) GroupMenusByCategory(mode string, perCategory int) (interface{}, error){
	// validasi
	if mode != "count" && mode != "list"{
		mode = "count"
	}

	// validasi per kategori
	if mode == "list" && perCategory < 1{
		perCategory = 10
	}
	if perCategory > 100{
		perCategory = 100
	}

	return s.repo.GroupByCategory(mode, perCategory)
}

// search
func (s *MenuService) SearchMenus(query string, page, perPage int) ([]models.Menu, *models.PaginationMeta, error) {
	if page < 1{
		page = 1
	}
	if perPage < 1{
		perPage = 10
	}
	if perPage > 100{
		perPage = 100
	}

	return s.repo.Search(query, page, perPage)
}

// cek menu by id
func (s *MenuService) ValidateMenuExists(id uint) error{
	_, err := s.repo.GetByID(id)
	if err != nil{
		if errors.Is(err, gorm.ErrRecordNotFound){
			return errors.New("menu tidak ditemukan")
		}
		return err
	}
	return nil
}