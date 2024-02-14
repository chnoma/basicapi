package main

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type Product struct {
	Id           int    `json:"id__c"`
	Manufacturer string `json:"manufacturer__c"`
	ModelNumber  string `json:"model_number__c"`
	Description  string `json:"description__c"`
	Qty          int    `json:"qty_available__c"`
	LeadTime     string `json:"leadtime__c"`
}

type Order struct {
	Id      int     `json:"id"`
	Qty     int     `json:"qty"`
	Status  string  `json:"status"`
	Product Product `json:"product"`
}

// ------------------------
// ORDER DATABASE API
// ------------------------

func (server *Server) get_all_orders() ([]Order, error) {
	row, err := server.Pg.Query(context.TODO(), `SELECT orders.id, orders.qty, order_statuses_enum.description,
                                                    products.id, products.manufacturer, products.model_number,
                                                    products.description, products.qty, products.lead_time FROM orders
                                                 JOIN products ON products.id = orders.product_id
                                                 JOIN order_statuses_enum ON orders.status_id = order_statuses_enum.id;`)
	if err != nil {
		return []Order{}, errors.New("failed to query database")
	}

	var orders []Order
	for row.Next() {
		var order Order
		err := row.Scan(&order.Id, &order.Qty, &order.Status,
			&order.Product.Id, &order.Product.Manufacturer, &order.Product.ModelNumber, &order.Product.Description,
			&order.Product.Qty, &order.Product.LeadTime)
		if err != nil {
			return []Order{}, errors.New("failed to scan Order from database")
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (server *Server) get_order_by_id(order_id int) (Order, error) {
	row := server.Pg.QueryRow(context.TODO(), `SELECT orders.id, orders.qty, order_statuses_enum.description,
                                                    products.id, products.manufacturer, products.model_number,
                                                    products.description, products.qty, products.lead_time FROM orders
                                                 JOIN products ON products.id = orders.product_id
                                                 JOIN order_statuses_enum ON orders.status_id = order_statuses_enum.id
                                                 WHERE orders.id = $1;`, order_id)
	var order Order
	err := row.Scan(&order.Id, &order.Qty, &order.Status,
		&order.Product.Id, &order.Product.Manufacturer, &order.Product.ModelNumber, &order.Product.Description,
		&order.Product.Qty, &order.Product.LeadTime)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Order{}, errors.New("order not found")
		}
		return Order{}, errors.New("failed to scan Order from database")
	}

	return order, nil
}

func (server *Server) place_order(product_id int, qty int) error {
	product, err := server.get_product_by_id(product_id)
	if err != nil {
		return errors.New("product not found")
	}

	if qty > product.Qty {
		return errors.New("quantity exceeds available product")
	}

	if qty < 1 {
		return errors.New("invalid order qty")
	}

	// Non-atomic, but does not matter for this
	command_tag, err := server.Pg.Exec(context.TODO(),
		`UPDATE products SET qty = $1 WHERE id = $2;`,
		product.Qty-qty, product.Id)

	if command_tag.RowsAffected() != 1 {
		return errors.New("failed to update product quantity")
	}

	command_tag, err = server.Pg.Exec(context.TODO(),
		`INSERT INTO orders (product_id, qty) VALUES ($1, $2);`,
		product.Id, qty)

	if command_tag.RowsAffected() != 1 {
		return errors.New("failed to create new order")
	}

	return nil
}

// ------------------------
// PRODUCT DATABASE API
// ------------------------

func (server *Server) get_all_products() ([]Product, error) {
	row, err := server.Pg.Query(context.TODO(), "SELECT * FROM products;")
	if err != nil {
		return []Product{}, errors.New("failed to query database")
	}

	var products []Product
	for row.Next() {
		var product Product
		err = row.Scan(&product.Id, &product.Manufacturer, &product.ModelNumber, &product.Description,
			&product.Qty, &product.LeadTime)
		if err != nil {
			return []Product{}, errors.New("failed to scan to Product from database")
		}

		products = append(products, product)
	}

	return products, nil
}

func (server *Server) get_product_by_id(product_id int) (Product, error) {
	row := server.Pg.QueryRow(context.TODO(), "SELECT * FROM products WHERE id=$1;", product_id)

	var product Product
	err := row.Scan(&product.Id, &product.Manufacturer, &product.ModelNumber, &product.Description,
		&product.Qty, &product.LeadTime)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Product{}, errors.New("product not found")
		}
		return Product{}, errors.New("failed to scan to Product from database")
	}

	return product, nil
}

func (server *Server) set_product_qty(product_id int, qty int) error {
	product, err := server.get_product_by_id(product_id)
	if err != nil {
		return errors.New("product not found")
	}

	command_tag, err := server.Pg.Exec(context.TODO(),
		`UPDATE products SET qty = $1 WHERE id = $2;`,
		qty, product.Id)

	if command_tag.RowsAffected() != 1 {
		return errors.New("failed to update product quantity")
	}

	return nil
}
