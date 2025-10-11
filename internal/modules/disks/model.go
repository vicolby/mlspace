package disks

import (
	"aispace/internal/services"
	"aispace/web/pages/disksweb"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Disk struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	Status    services.PVCStatus
	Owner     Owner
	Size      int  `db:"size"`
	Shared    bool `db:"shared"`
	Project   DiskProject
	CreatedAt time.Time `db:"created_at"`
}

type DiskProject struct {
	ID   uuid.UUID
	Name string
}

func (d *DiskProject) ToWebDiskProject(project DiskProject) disksweb.WebDiskProject {
	return disksweb.WebDiskProject{
		ID:   project.ID,
		Name: project.Name,
	}
}

func (d *Disk) GetNamespace() string {
	return fmt.Sprintf("project-%s", d.Project.ID.String())
}

func (d *Disk) GetPVCName() string {
	return fmt.Sprintf("disk-%s", d.ID.String())
}

func (d *Disk) GetPVCSize() string {
	return strconv.Itoa(d.Size)
}

type Owner struct {
	Username string `db:"name"`
	Email    string `db:"email"`
}

func (d *Disk) ToWebDisk(disk Disk) disksweb.WebDisk {
	return disksweb.WebDisk{
		ID:            disk.ID,
		Name:          disk.Name,
		Status:        disk.Status.String(),
		OwnerUsername: disk.Owner.Username,
		OwnerEmail:    disk.Owner.Email,
		Size:          disk.Size,
		Shared:        disk.Shared,
		Project:       disk.Project.ToWebDiskProject(disk.Project),
		CreatedAt:     disk.CreatedAt.Format("2006-01-02"),
	}

}
