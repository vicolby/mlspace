package disks

import (
	"aispace/internal/base"
	"aispace/internal/clients"
	"aispace/internal/consts"
	"aispace/web/pages/disksweb"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type DiskService struct {
	repository   DiskRepository
	kuberService *clients.KuberService
}

func NewDiskService(repository DiskRepository, kuberService *clients.KuberService) *DiskService {
	return &DiskService{repository: repository, kuberService: kuberService}
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

func (s *DiskService) CreateDisk(w http.ResponseWriter, r *http.Request) http.HandlerFunc {
	projectId := uuid.MustParse(r.FormValue("project_id"))
	ownerEmail := r.Context().Value(consts.ContextEmail).(string)
	ownerUsername := r.Context().Value(consts.ContextUsername).(string)
	diskId := uuid.New()

	diskName := r.FormValue("disk_name")
	diskSize, err := strconv.Atoi(r.FormValue("disk_size"))
	diskShared := r.FormValue("disk_shared") == "true"

	if ownerEmail == "" || err != nil {
		fmt.Printf("Invalid request: %s", err)
		return base.ErrorServe("Invalid request", http.StatusBadRequest, w)
	}

	projectName, err := s.repository.GetProjectNameByID(projectId)

	if err != nil {
		log.Printf("Error while fetching project name: %s", err)
		return base.ErrorServe("Something went wrong", http.StatusInternalServerError, w)
	}

	disk := Disk{
		ID:   diskId,
		Name: diskName,
		Owner: Owner{
			Username: ownerUsername,
			Email:    ownerEmail,
		},
		Size:   diskSize,
		Shared: diskShared,
		Project: DiskProject{
			ID:   projectId,
			Name: projectName,
		},
		CreatedAt: time.Now(),
	}

	err = s.repository.CreateDisk(disk)

	if err != nil {
		log.Printf("Error while creating disk: %s", err)
		return base.ErrorServe("Something went wrong", http.StatusInternalServerError, w)
	}

	webDisk := disk.ToWebDisk(disk)

	return base.Serve(disksweb.DiskRow(webDisk), w)
}

func (s *DiskService) DeleteDisk(w http.ResponseWriter, r *http.Request) http.HandlerFunc {
	diskId := uuid.MustParse(chi.URLParam(r, "disk_id"))

	err := s.repository.DeleteDisk(diskId)

	if err != nil {
		log.Printf("Error while deleting disk: %s", err)
		return base.ErrorServe("Something went wrong", http.StatusInternalServerError, w)
	}

	return base.ServeNoSwap(w)
}

func ProvideDiskService(repository DiskRepository, kuberService *clients.KuberService) *DiskService {
	return NewDiskService(repository, kuberService)
}
