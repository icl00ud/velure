package handlers

import (
	"strconv"
	"time"

	"product-service/internal/metrics"
	"product-service/internal/model"
	"product-service/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ProductHandler struct {
	service services.ProductService
}

const maxPageSize = 100

func NewProductHandler(service services.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
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

func (h *ProductHandler) GetProducts(c *fiber.Ctx) error {
	name := c.Query("name")
	if name == "" {
		name = c.Query("q")
	}

	if name != "" {
		products, err := h.service.GetProductsByName(c.Context(), name)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(products)
	}

	pageStr := c.Query("page")
	pageSizeStr := c.Query("limit")
	if pageSizeStr == "" {
		pageSizeStr = c.Query("pageSize")
	}
	category := c.Query("category")

	if pageStr == "" && pageSizeStr == "" {
		if category != "" {
			return fiber.NewError(fiber.StatusBadRequest, "category filter requires page and limit query parameters")
		}

		products, err := h.service.GetAllProducts(c.Context())
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(products)
	}

	if pageStr == "" || pageSizeStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "both page and limit query parameters are required")
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid page parameter")
	}
	if page < 1 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid page parameter: must be greater than or equal to 1")
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid limit parameter")
	}
	if pageSize < 1 || pageSize > maxPageSize {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid limit parameter: must be between 1 and 100")
	}

	metrics.ProductQueries.WithLabelValues("list").Inc()
	opStart := time.Now()

	var response *models.PaginatedProductsResponse
	if category != "" {
		response, err = h.service.GetProductsByPageAndCategory(c.Context(), page, pageSize, category)
	} else {
		response, err = h.service.GetProductsByPage(c.Context(), page, pageSize)
	}
	if err != nil {
		metrics.Errors.WithLabelValues("database").Inc()
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	metrics.ProductOperationDuration.WithLabelValues("list").Observe(time.Since(opStart).Seconds())
	metrics.SearchResultsReturned.Observe(float64(len(response.Products)))

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

func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Product ID is required")
	}

	return fiber.NewError(fiber.StatusNotImplemented, "product update is not implemented yet")
}

type updateInventoryRequest struct {
	QuantityChange int `json:"quantity_change"`
}

func (h *ProductHandler) PatchProductInventory(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Product ID is required")
	}

	var req updateInventoryRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	opStart := time.Now()
	if err := h.service.UpdateProductQuantity(c.Context(), id, req.QuantityChange); err != nil {
		metrics.Errors.WithLabelValues("inventory").Inc()
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	metrics.InventoryUpdates.WithLabelValues("success").Inc()
	metrics.ProductOperationDuration.WithLabelValues("update_quantity").Observe(time.Since(opStart).Seconds())

	return c.JSON(fiber.Map{
		"message": "Product quantity updated successfully",
	})
}
