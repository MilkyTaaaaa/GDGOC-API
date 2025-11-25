package handlers

import(
	"GDGOC-API/internal/models"
	"GDGOC-API/internal/services"
	"strconv"
	"strings"
	"github.com/gofiber/fiber/v2"
)

type MenuHandler struct{
	service *services.MenuService
}

// create instance baru MenuHandler
func NewMenuHandler(service *services.MenuService) *MenuHandler{
	return &MenuHandler{
		service: service,
	}
}

// POST /menu
func (h *MenuHandler) CreateMenu(c *fiber.Ctx) error{
	var req models.CreateMenuRequest

	//parsing
	if err := c.BodyParser(&req); err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Invalid request body",
			Errors: err.Error(),
		})
	}

	menu, err := h.service.CreateMenu(req)
	if err != nil{
		if strings.Contains(err.Error(), "validation") || strings.Contains(err.Error(), "required"){
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Message: "Validation failed",
				Errors: err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Message: "gagal membuat menu",
			Errors: err.Error(),
		})
	}

	// return sukses
	return c.Status(fiber.StatusCreated).JSON(models.MenuResponse{
		Message: "Menu berhasil dibuat",
		Data: *menu,
	})
}

// GET menu (filter & pagination)
func (h *MenuHandler) GetAllMenus(c *fiber.Ctx) error{
	//parsing
	filters := models.MenuFilters{
		Query:	c.Query("q"),
		Category:	c.Query("category"),
		MinPrice:	parseFloat(c.Query("min_price")),
		MaxPrice:	parseFloat(c.Query("max_price")),
		MaxCalories:	parseInt(c.Query("max_cal")),
		Page:	parseInt(c.Query("page")),
		PerPage: parseInt(c.Query("per_page")),
		Sort:	c.Query("sort"),
	}

	menus, pagination, err := h.service.GetAllMenus(filters)
	if err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Message: "Gagal mengambil data menu",
			Errors: err.Error(),
		})
	}

	// return
	response := models.MenuListResponse{
		Data:	menus,
		Pagination: pagination,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// Get menu by id
func (h *MenuHandler) GetMenuByID(c *fiber.Ctx) error{
	//parsing
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "ID menu invalid",
		})
	}

	menu, err := h.service.GetMenuByID(uint(id))
	if err != nil{
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "tidak ditemukan"){
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Message: "Menu tidak ditemukan",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Message: "Gagal retrieve menu",
			Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.MenuResponse{
		Data: *menu,
	})
}

// Update menu PUT
func (h *MenuHandler) UpdateMenu(c *fiber.Ctx) error{
	// parsing id
	id, err := strconv.ParseUint(c.Params("id"),10, 32)
	if err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "ID menu invalid",
		})
	}

	var req models.UpdateMenuRequest
	if err := c.BodyParser(&req); err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Invalid request body",
			Errors: err.Error(),
		})
	}

	menu, err := h.service.UpdateMenu(uint(id), req)
	if err != nil{
		if strings.Contains(err.Error(), "tidak ditemukan") || strings.Contains(err.Error(), "not found"){
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Message: "Menu tidak ditemukan",
			})
		}

		if strings.Contains(err.Error(), "validation") || strings.Contains(err.Error(), "required"){
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Message: "Validasi gagal",
				Errors: err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Message: "Gagal update menu",
			Errors: err.Error(),
		})
	}

	// return response
	return c.Status(fiber.StatusOK).JSON(models.MenuResponse{
		Message: "Menu berhasil diupdate",
		Data: *menu,
	})
}

// DELETE
func (h *MenuHandler) DeleteMenu(c *fiber.Ctx) error{
	// parsing ID
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "ID menu invalid",
		})
	}

	err = h.service.DeleteMenu(uint(id))
	if err != nil{
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "tidak ditemukan"){
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Message: "Menu tidak ditemukan",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Message: "Gagal menghapus menu",
			Errors: err.Error(),
		})
	}

	//return
	return c.Status(fiber.StatusOK).JSON(models.MessageResponse{
		Message: "Menu berhasil dihapus",
	})
}

// group by kategori
func (h *MenuHandler) GroupByCategory(c *fiber.Ctx) error{
	// parsing query parameter
	mode := c.Query("mode", "count")
	perCategory := parseInt(c.Query("per_category"))

	result, err := h.service.GroupMenusByCategory(mode, perCategory)
	if err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Message: "Gagal mengelompokkan menu",
			Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.GroupByCategoryResponse{
		Data: result,
	})
}

// search
func (h *MenuHandler) SearchMenus(c *fiber.Ctx) error{
	//parsing query parameter
	query := c.Query("q")
	page := parseInt(c.Query("page"))
	perPage := parseInt(c.Query("per_page"))

	menus, pagination, err := h.service.SearchMenus(query, page, perPage)
	if err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Message: "Gagal search menu",
			Errors: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.MenuListResponse{
		Data: menus,
		Pagination: pagination,
	})
}

// convert str -> int
func parseInt(s string) int{
	if s == ""{
		return 0
	}
	val, err := strconv.Atoi(s)
	if err != nil{
		return 0
	}
	return val
}

// convert str -> float64
func parseFloat(s string) float64{
	if s == ""{
		return 0
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}