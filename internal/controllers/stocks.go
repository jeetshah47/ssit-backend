package controllers

import (
	"net/http"

	"github.com/equitywala/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StockController struct {
	stockService *services.StockService
}

func NewStockController(stockService *services.StockService) *StockController {
	return &StockController{
		stockService: stockService,
	}
}

type CreateStockRequest struct {
	Name     string  `json:"name" binding:"required"`
	Symbol   string  `json:"symbol" binding:"required"`
	Exchange string  `json:"exchange" binding:"required"`
	Sector   *string `json:"sector"`
}

type UpdateStockRequest struct {
	Name     string  `json:"name" binding:"required"`
	Symbol   string  `json:"symbol" binding:"required"`
	Exchange string  `json:"exchange" binding:"required"`
	Sector   *string `json:"sector"`
	IsActive bool    `json:"isActive"`
}

func (c *StockController) Create(ctx *gin.Context) {
	var req CreateStockRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := services.CreateStockCmd{
		Name:     req.Name,
		Symbol:   req.Symbol,
		Exchange: req.Exchange,
		Sector:   req.Sector,
	}

	stock, err := c.stockService.Create(ctx.Request.Context(), cmd)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, stock)
}

func (c *StockController) GetAll(ctx *gin.Context) {
	stocks, err := c.stockService.GetAll(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, stocks)
}

func (c *StockController) GetByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	stock, err := c.stockService.GetByID(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, stock)
}

func (c *StockController) Update(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req UpdateStockRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := services.UpdateStockCmd{
		ID:       id,
		Name:     req.Name,
		Symbol:   req.Symbol,
		Exchange: req.Exchange,
		Sector:   req.Sector,
		IsActive: req.IsActive,
	}

	stock, err := c.stockService.Update(ctx.Request.Context(), cmd)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, stock)
}

func (c *StockController) Delete(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := c.stockService.Delete(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

