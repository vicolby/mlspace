package projects

type CreateProjectCommand struct {
	Name         string `validate:"required,min=3,max=100"`
	Description  string `validate:"required,min=10,max=500"`
	CPULimit     int    `validate:"required,gte=1"`
	RAMLimit     int    `validate:"required,gte=1"`
	StorageLimit int    `validate:"required,gte=1"`
}
