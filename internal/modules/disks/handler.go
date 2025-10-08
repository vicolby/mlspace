package disks

import "net/http"

type DiskHandler struct {
	diskService *DiskService
}

func NewDiskHandler(diskService *DiskService) *DiskHandler {
	return &DiskHandler{diskService: diskService}
}

func (h *DiskHandler) GetDisks(w http.ResponseWriter, r *http.Request) {
	handler := h.diskService.GetDisks(w, r)
	if handler != nil {
		handler(w, r)
	}
}

func (h *DiskHandler) GetProjectsForDisk(w http.ResponseWriter, r *http.Request) {
	handler := h.diskService.GetProjectsForDisk(w, r)
	if handler != nil {
		handler(w, r)
	}
}

func (h *DiskHandler) CreateDisk(w http.ResponseWriter, r *http.Request) {
	handler := h.diskService.CreateDisk(w, r)
	if handler != nil {
		handler(w, r)
	}
}

func ProvideDiskHandler(diskService *DiskService) *DiskHandler {
	return NewDiskHandler(diskService)
}
