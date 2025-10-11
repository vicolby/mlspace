package projects

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

type CreateProjectCommand struct {
	Name         string `validate:"required,min=3,max=100" form:"name"`
	Description  string `validate:"required,min=10,max=500" form:"description"`
	CPULimit     int    `validate:"required,gte=1" form:"cpu_limit"`
	RAMLimit     int    `validate:"required,gte=1" form:"ram_limit"`
	StorageLimit int    `validate:"required,gte=1" form:"storage_limit"`
}

func (c *CreateProjectCommand) Validate() error {
	err := validate.Struct(c)
	return err
}
