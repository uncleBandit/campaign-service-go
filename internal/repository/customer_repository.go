package repository

import (
	"database/sql"

	"github.com/unclebandit/smsleopard-backend/internal/model"
)

// CustomerRepositoryInterface defines methods used by service
type CustomerRepositoryInterface interface {
	GetByID(id int) (*model.Customer, error)
	ListAll() ([]model.Customer, error)
}

// CustomerRepository is the concrete implementation
type CustomerRepository struct {
	DB *sql.DB
}

// GetByID fetches a customer by ID
func (r *CustomerRepository) GetByID(id int) (*model.Customer, error) {
	query := `
        SELECT id, phone, first_name, last_name, location, preferred_product
        FROM customers
        WHERE id = $1
    `
	row := r.DB.QueryRow(query, id)

	var c model.Customer
	if err := row.Scan(&c.ID, &c.Phone, &c.FirstName, &c.LastName, &c.Location, &c.PreferredProduct); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // not found
		}
		return nil, err
	}
	return &c, nil
}

// ListAll fetches all customers (could be used for sending campaigns)
func (r *CustomerRepository) ListAll() ([]model.Customer, error) {
	query := `
        SELECT id, phone, first_name, last_name, location, preferred_product
        FROM customers
    `
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	customers := []model.Customer{}
	for rows.Next() {
		var c model.Customer
		if err := rows.Scan(&c.ID, &c.Phone, &c.FirstName, &c.LastName, &c.Location, &c.PreferredProduct); err != nil {
			return nil, err
		}
		customers = append(customers, c)
	}
	return customers, nil
}



