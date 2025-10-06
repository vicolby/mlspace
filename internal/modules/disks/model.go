package disks

import (
	"aispace/web/pages/disksweb"

	"github.com/google/uuid"
)

type Disk struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	Owner     Owner
	Size      int  `db:"size"`
	Shared    bool `db:"shared"`
	Project   string
	CreatedAt string `db:"create_at"`
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

type Owner struct {
	Username string `db:"name"`
	Email    string `db:"email"`
}

func (d *Disk) ToWebDisk(disk Disk) disksweb.WebDisk {
	return disksweb.WebDisk{
		ID:            disk.ID,
		Name:          disk.Name,
		OwnerUsername: disk.Owner.Username,
		OwnerEmail:    disk.Owner.Email,
		Size:          disk.Size,
		Shared:        disk.Shared,
		Project:       disk.Project,
		CreatedAt:     disk.CreatedAt,
	}

}
