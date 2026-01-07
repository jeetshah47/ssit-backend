package api

import (
	"github.com/go-playground/validator/v10"
)

// BindJSON binds JSON request body to struct and validates
func (c *Context) BindJSON(obj interface{}) error {
	if err := c.Context.ShouldBindJSON(obj); err != nil {
		return err
	}

	// Validate struct tags
	validate := validator.New()
	if err := validate.Struct(obj); err != nil {
		return err
	}

	return nil
}

// BindQuery binds query parameters to struct
func (c *Context) BindQuery(obj interface{}) error {
	return c.Context.ShouldBindQuery(obj)
}

// BindURI binds URI parameters to struct
func (c *Context) BindURI(obj interface{}) error {
	return c.Context.ShouldBindUri(obj)
}

// GetParam returns a URL parameter value
func (c *Context) GetParam(key string) string {
	return c.Context.Param(key)
}

// GetQuery returns a query parameter value
func (c *Context) GetQuery(key string) string {
	return c.Context.Query(key)
}

