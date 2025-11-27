#  Menu Catalog API

REST API untuk manajemen katalog menu restoran dengan fitur rekomendasi berbasis AI menggunakan Google Gemini AI.

##  Tech Stack

- **Backend Framework**: Go Fiber v2
- **Database**: PostgreSQL (Supabase)
- **ORM**: GORM
- **AI**: Google Gemini AI (gemini-2.0-flash)
- **Validation**: go-playground/validator
- **Language**: Go 1.24

##  Features

- ✅ **CRUD Operations** - Create, Read, Update, Delete menu
- ✅ **AI-Powered Recommendations** - Smart menu suggestions using Gemini AI
- ✅ **Advanced Search & Filtering** - Full-text search dengan multiple filters
- ✅ **Pagination** - Efficient data loading
- ✅ **Group by Category** - Organize menus by category
- ✅ **Dietary Filters** - Support untuk vegetarian, vegan, low-carb
- ✅ **Price & Calorie Filters** - Filter berdasarkan budget dan kesehatan

## 📦 Installation

### Prerequisites
- Go 1.21 or higher
- PostgreSQL database
- Google Gemini API Key

### Steps

1. **Clone repository:**
```bash
git clone https://github.com/MilkyTaaaaa/GDGOC-API.git
cd GDGOC-API
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Setup environment variables:**

Create `.env` file:
```env
PORT=3000
APP_ENV=development
DATABASE_URL=postgresql://user:password@host:port/database
GEMINI_API_KEY=your_gemini_api_key
TZ=Asia/Jakarta
```

4. **Run the application:**
```bash
go run cmd/main.go
```

Server akan berjalan di `http://localhost:3000`

## 📚 API Endpoints

### Health Check
```http
GET /health
```

### Menu Management

#### Create Menu
```http
POST /menu
Content-Type: application/json

{
  "name": "Nasi Goreng Pedas",
  "category": "foods",
  "price": 25000,
  "calories": 450,
  "ingredients": ["nasi", "cabai", "ayam", "telur"],
  "description": "Nasi goreng dengan level kepedasan tinggi"
}
```

#### Get All Menus (with filters & pagination)
```http
GET /menu?category=foods&max_price=30000&page=1&per_page=10
```

**Query Parameters:**
- `q` - Search query
- `category` - Filter by category (foods, drinks, desserts, snacks)
- `min_price` - Minimum price
- `max_price` - Maximum price
- `max_cal` - Maximum calories
- `page` - Page number (default: 1)
- `per_page` - Items per page (default: 10, max: 100)
- `sort` - Sort field and direction (e.g., `price:asc`, `name:desc`)

#### Get Menu by ID
```http
GET /menu/:id
```

#### Update Menu
```http
PUT /menu/:id
Content-Type: application/json

{
  "name": "Nasi Goreng Pedas Special",
  "category": "foods",
  "price": 27000,
  "calories": 450,
  "ingredients": ["nasi", "cabai", "ayam", "telur", "udang"],
  "description": "Nasi goreng pedas dengan udang"
}
```

#### Delete Menu
```http
DELETE /menu/:id
```

#### Search Menus
```http
GET /menu/search?q=pedas&page=1&per_page=10
```

#### Group by Category
```http
GET /menu/group-by-category?mode=count
```

**Modes:**
- `count` - Returns count per category
- `list` - Returns menu items grouped by category

### AI Recommendations 🤖

```http
POST /menu/recommendations
Content-Type: application/json

{
  "query": "saya ingin makanan pedas dan murah",
  "max_price": 30000,
  "diet": "vegetarian",
  "exclude": ["seafood", "daging"]
}
```

**Request Body:**
- `query` (required) - Natural language query
- `max_price` (optional) - Maximum price budget
- `diet` (optional) - Dietary preference (vegetarian, vegan, low-carb)
- `exclude` (optional) - Array of ingredients to exclude

**Response Example:**
```json
{
  "query": "saya ingin makanan pedas dan murah",
  "recommendations": [
    {
      "menu": {
        "id": 1,
        "name": "Nasi Goreng Pedas",
        "category": "foods",
        "price": 25000,
        "calories": 450,
        "ingredients": ["nasi", "cabai", "ayam", "telur"],
        "description": "Nasi goreng dengan level kepedasan tinggi"
      },
      "match_reason": "✅ Pedas sesuai permintaan Anda dan harga terjangkau"
    }
  ],
  "search_summary": "Ditemukan 3 menu yang cocok dengan 'saya ingin makanan pedas dan murah'",
  "suggestions": null
}
```

## 🏗️ Project Structure

```
GDGOC-API/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go           # Configuration management
│   ├── database/
│   │   └── database.go         # Database connection & setup
│   ├── models/
│   │   └── menu.go             # Data models & DTOs
│   ├── repositories/
│   │   └── menu_repo.go        # Data access layer
│   ├── services/
│   │   └── menu_service.go     # Business logic
│   ├── handlers/
│   │   └── menu_handler.go     # HTTP request handlers
│   ├── routes/
│   │   └── routes.go           # API route definitions
│   └── gemini/
│       ├── client.go           # Gemini AI client
│       ├── service.go          # AI recommendation logic
│       └── types.go            # Request/Response types
├── .env                        # Environment variables (not tracked)
├── .env.example                # Environment variables template
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## 🤖 AI Integration

Project ini menggunakan **Google Gemini AI** untuk memberikan rekomendasi menu yang personal dan intelligent.

### Cara Kerja:

1. **User Query Processing**
   - Natural language understanding
   - Extract preferences dari query

2. **Menu Filtering**
   - Apply dietary restrictions
   - Apply price limits
   - Exclude unwanted ingredients

3. **AI-Powered Ranking**
   - Gemini AI analyzes menu items
   - Generates personalized recommendations
   - Provides reasoning for each recommendation

4. **Fallback Mechanism**
   - Jika AI gagal, fallback ke keyword matching
   - Ensures service availability

## 📊 Database Schema

### Menu Table
```sql
CREATE TABLE menus (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    calories INTEGER,
    price DECIMAL(10,2) NOT NULL,
    ingredients TEXT[],
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_menus_category ON menus(category);
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `PORT` | Server port | `3000` |
| `APP_ENV` | Environment mode | `development` or `production` |
| `DATABASE_URL` | PostgreSQL connection string | `postgresql://user:pass@host:5432/db` |
| `GEMINI_API_KEY` | Google Gemini API key | `AIza...` |
| `TZ` | Timezone | `Asia/Jakarta` |

### Getting Gemini API Key

1. Visit [Google AI Studio](https://aistudio.google.com/app/apikey)
2. Create new API key
3. Copy and add to `.env` file

## 🧪 Testing

### Manual Testing with cURL

**Create Menu:**
```bash
curl -X POST http://localhost:3000/menu \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Nasi Goreng Pedas",
    "category": "foods",
    "price": 25000,
    "calories": 450,
    "ingredients": ["nasi", "cabai", "ayam", "telur"],
    "description": "Nasi goreng dengan level kepedasan tinggi"
  }'
```

**Get All Menus:**
```bash
curl http://localhost:3000/menu
```

**AI Recommendations:**
```bash
curl -X POST http://localhost:3000/menu/recommendations \
  -H "Content-Type: application/json" \
  -d '{
    "query": "saya ingin makanan pedas dan murah",
    "max_price": 30000
  }'
```

## 👨‍💻 Author

- GitHub: [@MilkyTaaaaa](https://github.com/MilkyTaaaaa)
- Email: orellsatrianitto@gmail.com

## 🙏 Acknowledgments

- [Go Fiber](https://gofiber.io/) - Web framework
- [GORM](https://gorm.io/) - ORM library
- [Google Gemini AI](https://ai.google.dev/) - AI recommendations
- [Supabase](https://supabase.com/) - PostgreSQL hosting

---
