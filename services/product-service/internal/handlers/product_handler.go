package handlers

import (
	"strconv"
	"time"

	"product-service/internal/metrics"
	"product-service/internal/models"
	"product-service/internal/services"

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

func (h *ProductHandler) GetProductsByPage(c *fiber.Ctx) error {
	start := time.Now()

	pageStr := c.Query("page")
	pageSizeStr := c.Query("pageSize")

	if pageStr == "" || pageSizeStr == "" {
		metrics.HTTPRequests.WithLabelValues("product-service", "GET", "/products", "400").Inc()
		return fiber.NewError(fiber.StatusBadRequest, "Missing query parameters")
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		metrics.HTTPRequests.WithLabelValues("product-service", "GET", "/products", "400").Inc()
		return fiber.NewError(fiber.StatusBadRequest, "Invalid page parameter")
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		metrics.HTTPRequests.WithLabelValues("product-service", "GET", "/products", "400").Inc()
		return fiber.NewError(fiber.StatusBadRequest, "Invalid pageSize parameter")
	}

	metrics.ProductQueries.WithLabelValues("list").Inc()
	opStart := time.Now()

	response, err := h.service.GetProductsByPage(c.Context(), page, pageSize)
	if err != nil {
		metrics.HTTPRequests.WithLabelValues("product-service", "GET", "/products", "500").Inc()
		metrics.Errors.WithLabelValues("database").Inc()
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	metrics.ProductOperationDuration.WithLabelValues("list").Observe(time.Since(opStart).Seconds())
	metrics.SearchResultsReturned.Observe(float64(len(response.Products)))
	metrics.HTTPRequests.WithLabelValues("product-service", "GET", "/products", "200").Inc()
	metrics.HTTPRequestDuration.WithLabelValues("product-service", "GET", "/products").Observe(time.Since(start).Seconds())

	return c.JSON(response)
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
	start := time.Now()

	var req models.UpdateQuantityRequest
	if err := c.BodyParser(&req); err != nil {
		metrics.HTTPRequests.WithLabelValues("product-service", "PUT", "/products/quantity", "400").Inc()
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.ProductID == "" {
		metrics.HTTPRequests.WithLabelValues("product-service", "PUT", "/products/quantity", "400").Inc()
		return fiber.NewError(fiber.StatusBadRequest, "Product ID is required")
	}

	opStart := time.Now()
	if err := h.service.UpdateProductQuantity(c.Context(), req.ProductID, req.QuantityChange); err != nil {
		metrics.HTTPRequests.WithLabelValues("product-service", "PUT", "/products/quantity", "400").Inc()
		metrics.Errors.WithLabelValues("inventory").Inc()
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	metrics.InventoryUpdates.WithLabelValues("success").Inc()
	metrics.ProductOperationDuration.WithLabelValues("update_quantity").Observe(time.Since(opStart).Seconds())
	metrics.HTTPRequests.WithLabelValues("product-service", "PUT", "/products/quantity", "200").Inc()
	metrics.HTTPRequestDuration.WithLabelValues("product-service", "PUT", "/products/quantity").Observe(time.Since(start).Seconds())

	return c.JSON(fiber.Map{
		"message": "Product quantity updated successfully",
	})
}
