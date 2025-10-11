package projects

import (
	"aispace/web/pages/projectsweb"
	"fmt"

	"github.com/google/uuid"
)

type Project struct {
	ID           uuid.UUID `db:"id"`
	Name         string    `db:"name"`
	Description  string    `db:"description"`
	Owner        Owner
	CPULimit     int    `db:"cpu_limit"`
	RAMLimit     int    `db:"ram_limit"`
	StorageLimit int    `db:"storage_limit"`
	CreatedAt    string `db:"created_at"`
}

func (p *Project) GetNamespace() string {
	return fmt.Sprintf("project-%s", p.ID.String())
}

func (p *Project) ToWebProject(project Project) projectsweb.WebProject {
	return projectsweb.WebProject{
		ID:            project.ID,
		Name:          project.Name,
		Description:   project.Description,
		OwnerUsername: project.Owner.Username,
		OwnerEmail:    project.Owner.Email,
		CPULimit:      project.CPULimit,
		RAMLimit:      project.RAMLimit,
		StorageLimit:  project.StorageLimit,
		CreatedAt:     project.CreatedAt,
	}

}

type Participant struct {
	ID       uuid.UUID `db:"id"`
	Username string    `db:"name"`
	Email    string    `db:"email"`
}

func (p *Participant) ToWebParticipant(participant Participant) projectsweb.WebProjectParticipant {
	return projectsweb.WebProjectParticipant{
		ID:    participant.ID,
		Name:  participant.Username,
		Email: participant.Email,
	}
}

type Owner struct {
	Username string `db:"name"`
	Email    string `db:"email"`
}
