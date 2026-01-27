package controllers

import (
	"github.com/equitywala/backend/internal/common/utils"
	"github.com/equitywala/backend/internal/services"
)

// PricingController handles pricing package endpoints
type PricingController struct {
	listPricingPackagesService *services.ListPricingPackagesService
}

// NewPricingController creates a new pricing controller
func NewPricingController(listPricingPackagesService *services.ListPricingPackagesService) *PricingController {
	return &PricingController{
		listPricingPackagesService: listPricingPackagesService,
	}
}

// ListPricingPackages handles listing all pricing packages
func (c *PricingController) ListPricingPackages(ctx *utils.Context) (interface{}, error) {
	packages, err := c.listPricingPackagesService.Execute(ctx.Request.Context())
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"packages": packages,
	}, nil
}

