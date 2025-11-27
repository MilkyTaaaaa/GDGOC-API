package handlers

import(
	"GDGOC-API/internal/gemini"
	"GDGOC-API/internal/models"
	"GDGOC-API/internal/services"
	"strconv"
	"strings"
	"fmt"
	"log"
	"github.com/gofiber/fiber/v2"
)

type MenuHandler struct{
	service *services.MenuService
	geminiService	*gemini.Service
}

// create instance baru MenuHandler
func NewMenuHandler(service *services.MenuService, geminiService *gemini.Service) *MenuHandler{
	return &MenuHandler{
		service: service,
		geminiService: geminiService,
	}
}

// mendapatkan rekomendasi menu
func (h *MenuHandler) GetRecommendations(c *fiber.Ctx) error {
    var req gemini.RecommendationReq
    
    // Parse request body
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
            Message: "Invalid request body",
            Errors:  err.Error(),
        })
    }

    // query tidak boleh kosong
    if strings.TrimSpace(req.Query) == "" {
        return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
            Message: "Query is required",
        })
    }

    // Dapatkan semua menu yang tersedia (dengan filter basic)
    filters := models.MenuFilters{
        MaxPrice: req.MaxPrice,
    }
    
    allMenus, _, err := h.service.GetAllMenus(filters)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
            Message: "Failed to get menus",
            Errors:  err.Error(),
        })
    }

    // Filter manual untuk diet
    filteredMenus := h.applyDietaryFilters(allMenus, req.Diet, req.Exclude)

    // Dapatkan rekomendasi dari Gemini AI
    var result *gemini.RecommendationResult
    
    if h.geminiService != nil {
        result, err = h.geminiService.GetRecommendations(req, filteredMenus)
        if err != nil {
            log.Printf("Gemini recommendation failed: %v", err)
            // Fallback ke basic recommendations
            result = h.getBasicRecommendations(req, filteredMenus)
        }
    } else {
        // menggunakan basic recommendations (jika Gemini tidak available)
        result = h.getBasicRecommendations(req, filteredMenus)
    }

    return c.Status(fiber.StatusOK).JSON(result)
}

// Filter menu berdasarkan dietary restrictions
func (h *MenuHandler) applyDietaryFilters(menus []models.Menu, diet string, exclude []string) []models.Menu {
    if diet == "" && len(exclude) == 0 {
        return menus 
    }

    var filtered []models.Menu
    
    for _, menu := range menus {
        // Filter berdasarkan diet
        if diet != "" && !h.matchesDiet(menu, diet) {
            continue
        }
        
        // Filter berdasarkan excluded ingredients
        if len(exclude) > 0 && h.containsExcluded(menu, exclude) {
            continue
        }
        
        filtered = append(filtered, menu)
    }
    
    return filtered
}

// mengecek apakah menu match dengan dietary requirement
func (h *MenuHandler) matchesDiet(menu models.Menu, diet string) bool {
    switch strings.ToLower(diet) {
    case "vegetarian":
        return !h.containsMeat(menu.Ingredients)
    case "vegan":
        return !h.containsAnimalProducts(menu.Ingredients)
    case "low-carb":
        return menu.Calories != nil && *menu.Calories < 400
    default:
        return true
    }
}

// mengecek apakah menu mengandung excluded ingredients
func (h *MenuHandler) containsExcluded(menu models.Menu, exclude []string) bool {
    for _, excluded := range exclude {
        for _, ingredient := range menu.Ingredients {
            if strings.Contains(strings.ToLower(ingredient), strings.ToLower(excluded)) {
                return true
            }
        }
    }
    return false
}

// Mengecek apakah mengandung daging (untuk vegetarian)
func (h *MenuHandler) containsMeat(ingredients []string) bool {
    meatKeywords := []string{"ayam", "daging", "sapi", "babi", "ikan", "udang", "cumi"}
    for _, ingredient := range ingredients {
        for _, meat := range meatKeywords {
            if strings.Contains(strings.ToLower(ingredient), meat) {
                return true
            }
        }
    }
    return false
}

// Mengecek apakah mengandung produk hewani (untuk vegan)
func (h *MenuHandler) containsAnimalProducts(ingredients []string) bool {
    animalProducts := []string{"susu", "keju", "telur", "madu", "mentega"}
    return h.containsMeat(ingredients) || h.containsAny(ingredients, animalProducts)
}

func (h *MenuHandler) containsAny(ingredients []string, keywords []string) bool {
    for _, ingredient := range ingredients {
        for _, keyword := range keywords {
            if strings.Contains(strings.ToLower(ingredient), keyword) {
                return true
            }
        }
    }
    return false
}

// Fallback recommendations tanpa AI
func (h *MenuHandler) getBasicRecommendations(req gemini.RecommendationReq, menus []models.Menu) *gemini.RecommendationResult {
    var recommendations []gemini.MenuRecommendation
    
    // Simple keyword matching
    queryLower := strings.ToLower(req.Query)
    
    for i, menu := range menus {
        if i >= 5 { // Maksimal 5 rekomendasi
            break
        }
        
        reason := " Rekomendasi berdasarkan preferensi Anda"
        menuText := strings.ToLower(menu.Name + " " + menu.Description)
        
        // Simple keyword matching untuk reason yang lebih spesifik
        if strings.Contains(queryLower, "pedas") && strings.Contains(menuText, "pedas") {
            reason = " Pedas sesuai permintaan Anda"
        } else if strings.Contains(queryLower, "sehat") && (menu.Calories != nil && *menu.Calories < 400) {
            reason = " Sehat dengan kalori terkontrol"
        } else if strings.Contains(queryLower, "murah") && menu.Price < 30000 {
            reason = " Harga terjangkau"
        }
        
        recommendations = append(recommendations, gemini.MenuRecommendation{
            Menu:       menu,
            MatchReason: reason,
        })
    }
    
    return &gemini.RecommendationResult{
        Query:          req.Query,
        Recommendations: recommendations,
        SearchSummary:  fmt.Sprintf("Ditemukan %d rekomendasi untuk '%s'", len(recommendations), req.Query),
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