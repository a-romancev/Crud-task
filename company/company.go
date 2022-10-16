package company

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrDuplicatedEntry = errors.New("duplicated company")
)

var Types = []string{"Corporations", "NonProfit", "Cooperative", "Sole Proprietorship"}

type Company struct {
	ID           uuid.UUID `json:"_id" bson:"_id"`
	Name         string    `json:"name" bson:"name"`
	Description  string    `json:"description" bson:"description"`
	EmployeesNum *int      `json:"employees_num" bson:"employees_num"`
	Registered   *bool     `json:"registered" bson:"registered"`
	Type         *string   `json:"type" bson:"type"`
}

func (c Company) Validate() error {
	if c.ID == uuid.Nil {
		return errors.New("id should not be empty")
	}
	if c.Name == "" {
		return errors.New("name cannot be empty")
	}
	if c.EmployeesNum == nil {
		return errors.New("employees_num cannot be empty")
	}
	if c.Registered == nil {
		return errors.New("registered cannot be empty")
	}
	if c.Type == nil {
		return errors.New("type cannot be empty")
	}

	if len(c.Name) > 15 {
		return errors.New("name cannot be more than 15 characters")
	}
	if len(c.Description) > 3000 {
		return errors.New("description cannot be more than 3000 characters")
	}
	if !checkIn(*c.Type, Types) {
		return errors.New("unknown company type")
	}
	return nil
}

type Lookup struct {
	ID uuid.UUID `json:"id"`
}

func checkIn(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
