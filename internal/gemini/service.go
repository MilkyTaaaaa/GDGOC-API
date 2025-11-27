package gemini

import (
    "context"
    "fmt"
    "strings"
    "time"
    "log"

    "GDGOC-API/internal/models"
    "github.com/google/generative-ai-go/genai"
)

// Logic business untuk Gemini
type Service struct {
    client *Client
    model  *genai.GenerativeModel
}

// Membuat instance baru Gemini service
func NewService(apiKey string) (*Service, error) {
    client, err := NewClient(apiKey)
    if err != nil {
        return nil, err
    }

    model := client.client.GenerativeModel("models/gemini-2.5-flash")
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    _, err = model.GenerateContent(ctx, genai.Text("test"))
    if err != nil {
        return nil, fmt.Errorf("model test failed: %v", err)
    }

    log.Println("Gemini service initialized")
    return &Service{
        client: client,
        model:  model,
    }, nil
}

// GetRecommendations - Mendapat rekomendasi menu by query pengguna
func (s *Service) GetRecommendations(req RecommendationReq, menus []models.Menu) (*RecommendationResult, error) {
    if len(menus) == 0 {
        return &RecommendationResult{
            Query:          req.Query,
            Recommendations: []MenuRecommendation{},
            SearchSummary:  "Tidak ada menu yang tersedia",
        }, nil
    }

    prompt := s.buildRecommendationPrompt(req, menus)
    
    ctx, cancel := context.WithTimeout(s.client.ctx, 15*time.Second)
    defer cancel()
    
    resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
    if err != nil {
        log.Printf("⚠️  Gemini AI failed: %v", err)
        return s.getBasicRecommendations(req, menus), nil
    }

    return s.parseGeminiResponse(req, menus, resp)
}

// parseGeminiResponse - Parse response dari package yang benar
func (s *Service) parseGeminiResponse(req RecommendationReq, menus []models.Menu, resp *genai.GenerateContentResponse) (*RecommendationResult, error) {
    if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
        return nil, fmt.Errorf("empty response from Gemini")
    }

    responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
    return s.extractRecommendationsFromText(req, menus, responseText)
}

// extractRecommendationsFromText - Extract rekomendasi dari text response
func (s *Service) extractRecommendationsFromText(req RecommendationReq, menus []models.Menu, responseText string) (*RecommendationResult, error) {
    log.Printf("🔍 Gemini Response: %s", responseText)
    
    var recommendations []MenuRecommendation
    lines := strings.Split(responseText, "\n")
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "- Menu:") || strings.HasPrefix(line, "Menu:") {
            parts := strings.Split(line, "Alasan:")
            if len(parts) == 2 {
                menuPart := strings.TrimPrefix(parts[0], "- Menu:")
                menuPart = strings.TrimPrefix(menuPart, "Menu:")
                menuName := strings.TrimSpace(menuPart)
                reason := strings.TrimSpace(parts[1])
                
                for _, menu := range menus {
                    if strings.Contains(strings.ToLower(menu.Name), strings.ToLower(menuName)) ||
                       strings.Contains(strings.ToLower(menuName), strings.ToLower(menu.Name)) {
                        recommendations = append(recommendations, MenuRecommendation{
                            Menu:       menu,
                            MatchReason: reason,
                        })
                        break
                    }
                }
            }
        }
    }

    // Fallback jika parsing gagal
    if len(recommendations) == 0 {
        recommendations = s.getFallbackRecommendations(req, menus)
    }
    
    return &RecommendationResult{
        Query:          req.Query,
        Recommendations: recommendations,
        SearchSummary:  s.generateSearchSummary(req, len(recommendations), len(menus)),
        Suggestions:    s.generateSuggestions(req, len(recommendations)),
    }, nil
}

// buildRecommendationPrompt - prompt untuk Gemini AI
func (s *Service) buildRecommendationPrompt(req RecommendationReq, menus []models.Menu) string {
    var menuStrings []string
    for i, menu := range menus {
        menuStr := fmt.Sprintf("%d. %s (Rp %.0f) - %s", 
            i+1, menu.Name, menu.Price, menu.Category)
        
        if menu.Calories != nil {
            menuStr += fmt.Sprintf(" - %d kalori", *menu.Calories)
        }
        
        if len(menu.Ingredients) > 0 {
            menuStr += fmt.Sprintf(" - Bahan: %s", strings.Join(menu.Ingredients, ", "))
        }
        
        menuStrings = append(menuStrings, menuStr)
    }

    return fmt.Sprintf(`ANDA ADALAH ASSISTANT AHLI REKOMENDASI MENU RESTORAN.

QUERY USER: "%s"
KRITERIA TAMBAHAN: %s

DAFTAR MENU YANG TERSEDIA:
%s

INSTRUKSI PENTING:
1. REKOMENDASIKAN HANYA MENU YANG BENAR-BENAR COCOK dengan query user
2. Jika query "minuman", REKOMENDASIKAN HANYA menu dengan kategori "drinks"
3. Jika query "makanan", REKOMENDASIKAN HANYA menu dengan kategori "foods" 
4. Jika query "dessert", REKOMENDASIKAN HANYA menu dengan kategori "desserts"
5. Jika query "snack", REKOMENDASIKAN HANYA menu dengan kategori "snacks"
6. BERI ALASAN SPESIFIK mengapa menu cocok dengan query
7. GUNAKAN ✅ untuk kelebihan, ⚠️ untuk kekurangan
8. MAXIMAL 5 REKOMENDASI saja
9. URUTKAN dari yang PALING COCOK

CONTOH UNTUK QUERY "minuman segar":
- Menu: Es Jeruk, Alasan: ✅ Minuman segar dari jeruk asli, ✅ Cocok untuk cuaca panas, ✅ Harga terjangkau

CONTOH UNTUK QUERY "makanan pedas":
- Menu: Nasi Goreng Pedas, Alasan: ✅ Pedas dari cabe segar, ✅ Gurih dengan bumbu khas, ⚠️ Kalori sedang

FORMAT OUTPUT:
- Menu: [nama_menu], Alasan: [alasan_singkat_dan_spesifik]

REKOMENDASI UNTUK "%s":`,
        req.Query,
        s.formatAdditionalCriteria(req),
        strings.Join(menuStrings, "\n"),
        req.Query,
    )
}
// formatAdditionalCriteria - format kriteria tambahan
func (s *Service) formatAdditionalCriteria(req RecommendationReq) string {
    var criteria []string
    
    if req.MaxPrice > 0 {
        criteria = append(criteria, fmt.Sprintf("maksimal Rp %.0f", req.MaxPrice))
    }
    if req.Diet != "" {
        criteria = append(criteria, req.Diet)
    }
    if len(req.Exclude) > 0 {
        criteria = append(criteria, fmt.Sprintf("hindari: %s", strings.Join(req.Exclude, ", ")))
    }

    if len(criteria) == 0 {
        return "tidak ada kriteria tambahan"
    }
    return strings.Join(criteria, ", ")
}

// getFallbackRecommendations - Fallback jika AI tidak memberi rekomendasi spesifik
func (s *Service) getFallbackRecommendations(req RecommendationReq, menus []models.Menu) []MenuRecommendation {
    var recommendations []MenuRecommendation
    
    // Simple keyword matching fallback
    queryLower := strings.ToLower(req.Query)
    
    for i, menu := range menus {
        if i >= 5 { // Maksimal 5 rekomendasi
            break
        }
        
        var reason string
        menuLower := strings.ToLower(menu.Name + " " + menu.Description)
        
        if strings.Contains(queryLower, "pedas") && strings.Contains(menuLower, "pedas") {
            reason = "✅ Pedas sesuai permintaan Anda"
        } else if strings.Contains(queryLower, "sehat") && (menu.Calories != nil && *menu.Calories < 500) {
            reason = "✅ Sehat dengan kalori terkontrol"
        } else if strings.Contains(queryLower, "murah") && menu.Price < 50000 {
            reason = "✅ Harga terjangkau"
        } else if strings.Contains(queryLower, "minuman") && menu.Category == "drinks" {
            reason = "✅ Minuman yang menyegarkan"
        } else if strings.Contains(queryLower, "makanan") && menu.Category == "foods" {
            reason = "✅ Makanan yang lezat"
        } else {
            reason = "✅ Rekomendasi terpopuler"
        }
        
        recommendations = append(recommendations, MenuRecommendation{
            Menu:       menu,
            MatchReason: reason,
        })
    }
    
    return recommendations
}

// getBasicRecommendations - Alias untuk getFallbackRecommendations
func (s *Service) getBasicRecommendations(req RecommendationReq, menus []models.Menu) *RecommendationResult {
    recommendations := s.getFallbackRecommendations(req, menus)
    return &RecommendationResult{
        Query:          req.Query,
        Recommendations: recommendations,
        SearchSummary:  s.generateSearchSummary(req, len(recommendations), len(menus)),
        Suggestions:    s.generateSuggestions(req, len(recommendations)),
    }
}

// generateSearchSummary - Generate summary
func (s *Service) generateSearchSummary(req RecommendationReq, found int, total int) string {
    if found == 0 {
        return fmt.Sprintf("Tidak ditemukan menu yang cocok dengan '%s'", req.Query)
    }
    return fmt.Sprintf("Ditemukan %d menu yang cocok dengan '%s'", found, req.Query)
}

// generateSuggestions - Generate suggestions untuk user
func (s *Service) generateSuggestions(req RecommendationReq, found int) []string {
    if found == 0 {
        return []string{
            "Coba gunakan kata kunci yang lebih spesifik",
            "Lihat menu populer di kategori Foods",
        }
    }
    
    if found < 3 {
        return []string{
            "Coba tambahkan kata 'ayam' atau 'seafood' untuk lebih banyak pilihan",
        }
    }

    return nil
}

// Close - Close koneksi
func (s *Service) Close() error {
    if s.client != nil {
        return s.client.Close()
    }
    return nil
}