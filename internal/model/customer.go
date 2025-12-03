// internal/model/customer.go
package model

type Customer struct {
    ID               int    `db:"id" json:"id"`
    Phone            string `db:"phone" json:"phone"`
    FirstName        string `db:"first_name" json:"first_name"`
    LastName         string `db:"last_name" json:"last_name"`
    Location         string `db:"location" json:"location"`
    PreferredProduct string `db:"preferred_product" json:"preferred_product"`
}
