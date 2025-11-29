package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"product-service/internal/config"
	"product-service/internal/models"
	"product-service/internal/repository"

	"github.com/joho/godotenv"
)

type ProductData struct {
	Name        string
	Description string
	Price       float64
	Brand       string
	Category    string
	SKU         string
}

type seedDependencies struct {
	loadEnv   func() error
	buildRepo func(cfg *config.Config) (repository.ProductRepository, func(context.Context) error, func(), error)
	generate  func() []models.CreateProductRequest
}

var defaultSeedDeps = seedDependencies{
	loadEnv: func() error { return godotenv.Load() },
	buildRepo: func(cfg *config.Config) (repository.ProductRepository, func(context.Context) error, func(), error) {
		mongodb, err := config.NewMongoDB(cfg.MongoURI)
		if err != nil {
			return nil, nil, nil, err
		}

		redis := config.NewRedis(cfg.RedisAddr, cfg.RedisPassword)
		repo := repository.NewProductRepository(mongodb.Database(cfg.DatabaseName), redis)

		return repo, mongodb.Disconnect, func() {
			_ = redis.Close()
		}, nil
	},
	generate: generatePetProducts,
}

var seedFatalf = log.Fatal

func main() {
	if err := runSeed(defaultSeedDeps); err != nil {
		seedFatalf(err)
	}
}

func runSeed(deps seedDependencies) error {
	// Load environment variables
	if err := deps.loadEnv(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.New()

	log.Printf("Connecting to MongoDB: %s", cfg.DatabaseName)

	repo, mongoDisconnect, redisClose, err := deps.buildRepo(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if mongoDisconnect != nil {
		defer func() {
			if err := mongoDisconnect(context.Background()); err != nil {
				log.Printf("Error disconnecting from MongoDB: %v", err)
			}
		}()
	}

	if redisClose != nil {
		defer redisClose()
	}

	// Generate and insert products
	products := deps.generate()

	log.Printf("Inserting %d products into MongoDB...", len(products))

	ctx := context.Background()
	successCount := 0
	for i, product := range products {
		_, err := repo.CreateProduct(ctx, product)
		if err != nil {
			log.Printf("Error creating product %d (%s): %v", i+1, product.Name, err)
		} else {
			successCount++
			if successCount%5 == 0 {
				log.Printf("Inserted %d/%d products...", successCount, len(products))
			}
		}
	}

	log.Printf("✅ Successfully inserted %d out of %d products!", successCount, len(products))
	return nil
}

func generatePetProducts() []models.CreateProductRequest {
	rand.Seed(time.Now().UnixNano())

	// Produtos de cães
	dogProducts := []ProductData{
		{
			Name:        "Ração Premium Cães Adultos Frango e Arroz",
			Description: "Ração super premium para cães adultos com frango real e arroz integral. Rico em proteínas e nutrientes essenciais para a saúde do seu pet.",
			Price:       89.99,
			Brand:       "Royal Canin",
			Category:    "Alimentação",
			SKU:         "RC-DOG-ADULT-3KG",
		},
		{
			Name:        "Ração Natural Cães Filhotes Salmão",
			Description: "Alimento natural para filhotes com salmão fresco, quinoa e vegetais. Sem conservantes artificiais, ideal para o desenvolvimento saudável.",
			Price:       125.50,
			Brand:       "Farmina N&D",
			Category:    "Alimentação",
			SKU:         "FN-PUPPY-SAL-2.5KG",
		},
		{
			Name:        "Petisco Natural Osso de Couro",
			Description: "Osso de couro natural 100% bovino, ideal para a higiene dental e entretenimento. Longa duração e sabor irresistível.",
			Price:       15.99,
			Brand:       "DogChew",
			Category:    "Petiscos",
			SKU:         "DC-BONE-NATURAL",
		},
		{
			Name:        "Bola Interativa com Dispenser de Petiscos",
			Description: "Bola de borracha resistente com compartimento para petiscos. Estimula o exercício mental e físico do seu cão.",
			Price:       45.90,
			Brand:       "Kong",
			Category:    "Brinquedos",
			SKU:         "KONG-BALL-TREAT",
		},
		{
			Name:        "Corda Dental Tri-Nó Algodão",
			Description: "Brinquedo de corda de algodão natural com três nós. Ajuda na limpeza dos dentes e fortalece a mandíbula.",
			Price:       24.99,
			Brand:       "PetPlay",
			Category:    "Brinquedos",
			SKU:         "PP-ROPE-3KNOT",
		},
		{
			Name:        "Coleira Ajustável Couro Legítimo",
			Description: "Coleira de couro legítimo com fivela de metal resistente. Confortável e durável para uso diário.",
			Price:       67.50,
			Brand:       "LeatherPet",
			Category:    "Acessórios",
			SKU:         "LP-COLLAR-LEATHER-M",
		},
		{
			Name:        "Cama Ortopédica Memory Foam",
			Description: "Cama ortopédica com espuma memory foam para máximo conforto. Capa removível e lavável.",
			Price:       189.90,
			Brand:       "ComfortPet",
			Category:    "Camas e Descanso",
			SKU:         "CP-BED-ORTHO-L",
		},
	}

	// Produtos de gatos
	catProducts := []ProductData{
		{
			Name:        "Ração Premium Gatos Castrados Frango",
			Description: "Ração especial para gatos castrados com controle de peso. Rica em proteínas e fibras, baixo teor de gordura.",
			Price:       75.99,
			Brand:       "Hill's",
			Category:    "Alimentação",
			SKU:         "HILLS-CAT-CAST-1.5KG",
		},
		{
			Name:        "Sachê Gourmet Peixe ao Molho",
			Description: "Alimento úmido gourmet com pedaços de peixe em molho saboroso. Rico em nutrientes e irresistível.",
			Price:       3.50,
			Brand:       "Whiskas",
			Category:    "Alimentação",
			SKU:         "WK-SACHET-FISH",
		},
		{
			Name:        "Arranhador Torre com Brinquedos",
			Description: "Torre arranhadora de sisal com plataformas e brinquedos suspensos. Ideal para exercício e diversão.",
			Price:       159.90,
			Brand:       "CatTree",
			Category:    "Brinquedos",
			SKU:         "CT-TOWER-SCRATCH",
		},
		{
			Name:        "Varinha com Penas Naturais",
			Description: "Brinquedo interativo com penas naturais coloridas. Estimula o instinto de caça dos felinos.",
			Price:       19.99,
			Brand:       "FelinePlay",
			Category:    "Brinquedos",
			SKU:         "FP-WAND-FEATHER",
		},
		{
			Name:        "Caixa de Areia Fechada com Filtro",
			Description: "Caixa de areia fechada com sistema de filtro de odores. Fácil limpeza e máxima higiene.",
			Price:       129.90,
			Brand:       "CleanCat",
			Category:    "Higiene",
			SKU:         "CC-LITTER-CLOSED",
		},
	}

	// Produtos de pássaros
	birdProducts := []ProductData{
		{
			Name:        "Mistura de Sementes Premium Canários",
			Description: "Mistura especial de sementes selecionadas para canários. Rica em nutrientes e vitaminas essenciais.",
			Price:       28.90,
			Brand:       "BirdSeed",
			Category:    "Alimentação",
			SKU:         "BS-CANARY-MIX-1KG",
		},
		{
			Name:        "Gaiola Spaciosa com Poleiros",
			Description: "Gaiola espaçosa com poleiros de madeira natural e comedouros em inox. Design moderno e funcional.",
			Price:       299.90,
			Brand:       "BirdHome",
			Category:    "Gaiolas",
			SKU:         "BH-CAGE-LARGE",
		},
	}

	// Produtos de peixes
	fishProducts := []ProductData{
		{
			Name:        "Ração em Flocos Peixes Tropicais",
			Description: "Ração balanceada em flocos para peixes tropicais. Rica em proteínas e vitaminas para cores vibrantes.",
			Price:       18.90,
			Brand:       "AquaFood",
			Category:    "Alimentação",
			SKU:         "AF-FLAKES-TROPICAL",
		},
		{
			Name:        "Filtro Submerso para Aquários",
			Description: "Filtro submerso silencioso com bomba integrada. Ideal para aquários de 50 a 100 litros.",
			Price:       89.50,
			Brand:       "AquaTech",
			Category:    "Aquários",
			SKU:         "AT-FILTER-SUB-100L",
		},
	}

	// Combinar todos os produtos
	allProducts := append(dogProducts, catProducts...)
	allProducts = append(allProducts, birdProducts...)
	allProducts = append(allProducts, fishProducts...)

	// Converter para CreateProductRequest
	var requests []models.CreateProductRequest
	for _, p := range allProducts {
		request := models.CreateProductRequest{
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Rating:      randomRating(),
			Category:    p.Category,
			Quantity:    randomQuantity(),
			Images:      generateImages(p.Category, p.Name),
			Dimensions:  generateDimensions(p.Category),
			Brand:       p.Brand,
			Colors:      generateColors(p.Category),
			SKU:         p.SKU,
		}
		requests = append(requests, request)
	}

	return requests
}

func randomRating() float64 {
	// Rating entre 3.5 e 5.0
	return 3.5 + rand.Float64()*1.5
}

func randomQuantity() int {
	// Quantidade entre 10 e 110
	return rand.Intn(100) + 10
}

func generateImages(category, productName string) []string {
	baseUrls := map[string]string{
		"Alimentação":      "https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=400&h=300&fit=crop",
		"Brinquedos":       "https://images.unsplash.com/photo-1601758228041-f3b2795255f1?w=400&h=300&fit=crop",
		"Acessórios":       "https://images.unsplash.com/photo-1583337130417-3346a1be7dee?w=400&h=300&fit=crop",
		"Camas e Descanso": "https://images.unsplash.com/photo-1583337130417-3346a1be7dee?w=400&h=300&fit=crop",
		"Petiscos":         "https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=400&h=300&fit=crop",
		"Higiene":          "https://images.unsplash.com/photo-1545249390-6bdfa286032f?w=400&h=300&fit=crop",
		"Gaiolas":          "https://images.unsplash.com/photo-1555169062-013468b47731?w=400&h=300&fit=crop",
		"Aquários":         "https://images.unsplash.com/photo-1554263897-4bfa012dcac0?w=400&h=300&fit=crop",
	}

	baseUrl := baseUrls[category]
	if baseUrl == "" {
		baseUrl = "https://images.unsplash.com/photo-1589924691995-400dc9ecc119?w=400&h=300&fit=crop"
	}

	images := []string{
		baseUrl + "&seed=" + randomString(8),
		baseUrl + "&seed=" + randomString(8),
		baseUrl + "&seed=" + randomString(8),
	}

	return images
}

func generateDimensions(category string) models.Dimensions {
	dimensionsMap := map[string]models.Dimensions{
		"Alimentação":      {Height: 25, Width: 15, Length: 10, Weight: 2.0},
		"Brinquedos":       {Height: 10, Width: 10, Length: 15, Weight: 0.3},
		"Acessórios":       {Height: 5, Width: 20, Length: 30, Weight: 0.5},
		"Camas e Descanso": {Height: 15, Width: 60, Length: 80, Weight: 2.5},
		"Petiscos":         {Height: 15, Width: 10, Length: 8, Weight: 0.5},
		"Higiene":          {Height: 30, Width: 40, Length: 50, Weight: 1.8},
		"Gaiolas":          {Height: 50, Width: 40, Length: 30, Weight: 5.0},
		"Aquários":         {Height: 15, Width: 30, Length: 25, Weight: 1.2},
	}

	dims, exists := dimensionsMap[category]
	if !exists {
		return models.Dimensions{Height: 10, Width: 10, Length: 10, Weight: 1.0}
	}
	return dims
}

func generateColors(category string) []string {
	colorMap := map[string][]string{
		"Alimentação":      {"Natural", "Marrom"},
		"Brinquedos":       {"Vermelho", "Azul", "Verde", "Amarelo", "Rosa"},
		"Acessórios":       {"Preto", "Marrom", "Azul", "Vermelho"},
		"Camas e Descanso": {"Cinza", "Bege", "Azul", "Marrom"},
		"Petiscos":         {"Natural", "Marrom"},
		"Higiene":          {"Branco", "Cinza", "Azul"},
		"Gaiolas":          {"Branco", "Preto", "Prata"},
		"Aquários":         {"Transparente", "Azul"},
	}

	colors, exists := colorMap[category]
	if !exists {
		return []string{"Variadas"}
	}

	// Retornar 1-3 cores aleatórias
	numColors := rand.Intn(3) + 1
	if numColors > len(colors) {
		numColors = len(colors)
	}

	return colors[:numColors]
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
