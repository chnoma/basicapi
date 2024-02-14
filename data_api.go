package main

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type OrderRequest struct {
	ProductId int `json:"product_id"`
	Qty       int `json:"qty"`
}

// ------------------------
// ORDER JSON DATA ROUTES
// ------------------------

func (server *Server) setup_order_data_routes(e *echo.Echo) {
	e.GET("/api/orders/", server.handle_get_orders)
	e.GET("/api/orders/:id", server.handle_get_order_by_id)
	e.POST("/api/orders/", server.handle_post_new_order)
}

// GET /orders/
func (server *Server) handle_get_orders(c echo.Context) error {
	orders, err := server.get_all_orders()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{err.Error()})
	}
	return c.JSON(http.StatusOK, orders)
}

// GET /orders/:id
func (server *Server) handle_get_order_by_id(c echo.Context) error {
	order_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{"Invalid ID Format"})
	}
	order, err := server.get_order_by_id(order_id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{err.Error()})
	}
	return c.JSON(http.StatusOK, order)
}

// POST /orders/
func (server *Server) handle_post_new_order(c echo.Context) error {
	var request OrderRequest
	// IMPORTANT!!
	// Binding values like this does not perform json validation
	// e.g., if {"product_id": 1} is supplied, it will assume qty to be zero, the default int value
	// This is ignored in this POC
	err := c.Bind(&request)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{"Invalid request format"})
	}

	err = server.place_order(request.ProductId, request.Qty)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{err.Error()})
	}

	return c.JSON(http.StatusOK, `{"status":"accepted"}`)
}

// ------------------------
// PRODUCT JSON DATA ROUTES
// ------------------------

func (server *Server) setup_product_data_routes(c *echo.Echo) {
	c.GET("/api/products/", server.handle_get_products)
	c.GET("/api/products/:id", server.handle_get_product_by_id)
	c.POST("/api/products/", server.handle_post_product_qty)
}

// GET /products/
func (server *Server) handle_get_products(c echo.Context) error {
	products, err := server.get_all_products()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{"Internal Database Error"})
	}
	return c.JSON(http.StatusOK, products)
}

// GET /products/:id
func (server *Server) handle_get_product_by_id(c echo.Context) error {
	product_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{"Invalid ID Format"})
	}

	product, err := server.get_product_by_id(product_id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{err.Error()})
	}

	return c.JSON(http.StatusOK, product)
}

// POST /products/
func (server *Server) handle_post_product_qty(c echo.Context) error {
	var request OrderRequest
	// IMPORTANT!!
	// Binding values like this does not perform json validation
	// e.g., if {"product_id": 1} is supplied, it will assume qty to be zero, the default int value
	err := c.Bind(&request)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{"Invalid request format"})
	}

	err = server.set_product_qty(request.ProductId, request.Qty)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{err.Error()})
	}

	return c.JSON(http.StatusOK, `{"status":"accepted"}`)
}
