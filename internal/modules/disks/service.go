package disks

import (
	"aispace/internal/base"
	"aispace/web/pages/disksweb"
	"log"
	"net/http"
)

type DiskService struct {
	repository DiskRepository
}

func NewDiskService(repository DiskRepository) *DiskService {
	return &DiskService{repository: repository}
}

func (s *DiskService) GetDisks(w http.ResponseWriter, r *http.Request) http.HandlerFunc {
	disks, err := s.repository.GetDisks(r.Context())

	if err != nil {
		log.Printf("Error while fetching disks: %s", err)
		return base.ErrorServe("Something went wrong", http.StatusInternalServerError, w)
	}

	var webDiskList []disksweb.WebDisk

	for _, disk := range disks {
		webDiskList = append(webDiskList, disk.ToWebDisk(disk))
	}

	if r.Header.Get("HX-Request") == "true" {
		return base.Serve(disksweb.DisksPartial(webDiskList), w)
	}
	return base.Serve(disksweb.DisksFull(webDiskList), w)
}

func (s *DiskService) GetProjectsForDisk(w http.ResponseWriter, r *http.Request) http.HandlerFunc {
	name := r.URL.Query().Get("project_name")
	projects, err := s.repository.GetProjectsByName(r.Context(), name)

	if err != nil {
		log.Printf("Error while fetching projects: %s", err)
		return base.ErrorServe("Something went wrong", http.StatusInternalServerError, w)
	}

	var webDiskProjectsList []disksweb.WebDiskProject

	for _, project := range projects {
		webDiskProjectsList = append(webDiskProjectsList, project.ToWebDiskProject(project))
	}

	return base.Serve(disksweb.DiskProjects(webDiskProjectsList), w)
}

func ProvideDiskService(repository DiskRepository) *DiskService {
	return NewDiskService(repository)
}
