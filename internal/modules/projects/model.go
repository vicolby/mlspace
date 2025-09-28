package projects

type Project struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	OwnerUsername string `json:"owner_username"`
	CPULimit      int    `json:"cpu_limit"`
	RAMLimit      int    `json:"ram_limit"`
	StorageLimit  int    `json:"storage_limit"`
}
