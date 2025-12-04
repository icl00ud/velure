package handlers

import (
	"strconv"
	"time"

	"github.com/icl00ud/velure-shared/logger"
	"product-service/internal/metrics"
	"product-service/internal/model"
	"product-service/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ProductHandler struct {
	service services.ProductService
}

func NewProductHandler(service services.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

func (h *ProductHandler) GetAllProducts(c *fiber.Ctx) error {
	products, err := h.service.GetAllProducts(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(products)
}

func (h *ProductHandler) GetProductById(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Product ID is required")
	}

	product, err := h.service.GetProductById(c.Context(), id)
	if err != nil {
		if err.Error() == "product not found" {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(product)
}

func (h *ProductHandler) GetProductsByName(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Product name is required")
	}

	products, err := h.service.GetProductsByName(c.Context(), name)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(products)
}

// GetProductsREST is a REST-style endpoint that accepts both 'limit' and 'pageSize' query parameters
func (h *ProductHandler) GetProductsREST(c *fiber.Ctx) error {
	pageStr := c.Query("page")
	// Accept both 'limit' (REST-style) and 'pageSize' (legacy) - prefer 'limit'
	pageSizeStr := c.Query("limit")
	if pageSizeStr == "" {
		pageSizeStr = c.Query("pageSize")
	}

	if pageStr == "" || pageSizeStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing query parameters (page and limit/pageSize required)")
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid page parameter")
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid limit/pageSize parameter")
	}

	metrics.ProductQueries.WithLabelValues("list").Inc()
	opStart := time.Now()

	response, err := h.service.GetProductsByPage(c.Context(), page, pageSize)
	if err != nil {
		metrics.Errors.WithLabelValues("database").Inc()
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	metrics.ProductOperationDuration.WithLabelValues("list").Observe(time.Since(opStart).Seconds())
	metrics.SearchResultsReturned.Observe(float64(len(response.Products)))

	return c.JSON(response)
}

// GetProductsByPage is the legacy endpoint that requires 'pageSize'
func (h *ProductHandler) GetProductsByPage(c *fiber.Ctx) error {
	pageStr := c.Query("page")
	pageSizeStr := c.Query("pageSize")

	if pageStr == "" || pageSizeStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing query parameters")
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid page parameter")
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid pageSize parameter")
	}

	metrics.ProductQueries.WithLabelValues("list").Inc()
	opStart := time.Now()

	response, err := h.service.GetProductsByPage(c.Context(), page, pageSize)
	if err != nil {
		metrics.Errors.WithLabelValues("database").Inc()
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	metrics.ProductOperationDuration.WithLabelValues("list").Observe(time.Since(opStart).Seconds())
	metrics.SearchResultsReturned.Observe(float64(len(response.Products)))

	return c.JSON(response)
}

// SearchProducts is a REST-style search endpoint that accepts 'q' query parameter
func (h *ProductHandler) SearchProducts(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Search query parameter 'q' is required")
	}

	products, err := h.service.GetProductsByName(c.Context(), query)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(products)
}

func (h *ProductHandler) GetProductsByPageAndCategory(c *fiber.Ctx) error {
	pageStr := c.Query("page")
	pageSizeStr := c.Query("pageSize")
	category := c.Query("category")

	if pageStr == "" || pageSizeStr == "" || category == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing query parameters")
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid page parameter")
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid pageSize parameter")
	}

	response, err := h.service.GetProductsByPageAndCategory(c.Context(), page, pageSize, category)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(response)
}

func (h *ProductHandler) GetProductsCount(c *fiber.Ctx) error {
	count, err := h.service.GetProductsCount(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(count)
}

func (h *ProductHandler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.service.GetCategories(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(categories)
}

func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var req models.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	product, err := h.service.CreateProduct(c.Context(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(product)
}

func (h *ProductHandler) DeleteProductsByName(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Product name is required")
	}

	if err := h.service.DeleteProductsByName(c.Context(), name); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ProductHandler) DeleteProductById(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Product ID is required")
	}

	if err := h.service.DeleteProductById(c.Context(), id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ProductHandler) UpdateProductQuantity(c *fiber.Ctx) error {
	logger.Debug("UpdateProductQuantity received", logger.String("body", string(c.Body())))

	var req models.UpdateQuantityRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	logger.Debug("Parsed UpdateQuantityRequest", logger.String("product_id", req.ProductID), logger.Int("quantity_change", req.QuantityChange))

	if req.ProductID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Product ID is required")
	}

	opStart := time.Now()
	if err := h.service.UpdateProductQuantity(c.Context(), req.ProductID, req.QuantityChange); err != nil {
		metrics.Errors.WithLabelValues("inventory").Inc()
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	metrics.InventoryUpdates.WithLabelValues("success").Inc()
	metrics.ProductOperationDuration.WithLabelValues("update_quantity").Observe(time.Since(opStart).Seconds())

	return c.JSON(fiber.Map{
		"message": "Product quantity updated successfully",
	})
}
