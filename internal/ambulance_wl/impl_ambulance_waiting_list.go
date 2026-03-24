package ambulance_wl

import (
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type implAmbulanceWaitingListAPI struct{}

func NewAmbulanceWaitingListApi() AmbulanceWaitingListAPI {
	return &implAmbulanceWaitingListAPI{}
}

// CreateWaitingListEntry - Saves new entry into waiting list
func (o *implAmbulanceWaitingListAPI) CreateWaitingListEntry(c *gin.Context) {
	updateAmbulanceFunc(c, func(c *gin.Context, ambulance *Ambulance) (*Ambulance, interface{}, int) {
		var entry WaitingListEntry

		if err := c.ShouldBindJSON(&entry); err != nil {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request body",
				"error":   err.Error(),
			}, http.StatusBadRequest
		}

		if entry.PatientId == "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Patient ID is required",
			}, http.StatusBadRequest
		}

		if entry.Id == "" || entry.Id == "@new" {
			entry.Id = uuid.NewString()
		}

		conflictIndx := slices.IndexFunc(ambulance.WaitingList, func(waiting WaitingListEntry) bool {
			return entry.Id == waiting.Id || entry.PatientId == waiting.PatientId
		})
		if conflictIndx >= 0 {
			return nil, gin.H{
				"status":  http.StatusConflict,
				"message": "Entry already exists",
			}, http.StatusConflict
		}

		ambulance.WaitingList = append(ambulance.WaitingList, entry)
		ambulance.reconcileWaitingList()

		entryIndx := slices.IndexFunc(ambulance.WaitingList, func(waiting WaitingListEntry) bool {
			return entry.Id == waiting.Id
		})
		if entryIndx < 0 {
			return nil, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to save entry",
			}, http.StatusInternalServerError
		}
		return ambulance, ambulance.WaitingList[entryIndx], http.StatusOK
	})
}

// DeleteWaitingListEntry - Deletes specific entry
func (o *implAmbulanceWaitingListAPI) DeleteWaitingListEntry(c *gin.Context) {
	updateAmbulanceFunc(c, func(c *gin.Context, ambulance *Ambulance) (*Ambulance, interface{}, int) {
		entryID := c.Param("entryId")
		if entryID == "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Entry ID is required",
			}, http.StatusBadRequest
		}

		entryIndx := slices.IndexFunc(ambulance.WaitingList, func(waiting WaitingListEntry) bool {
			return entryID == waiting.Id
		})
		if entryIndx < 0 {
			return nil, gin.H{
				"status":  http.StatusNotFound,
				"message": "Entry not found",
			}, http.StatusNotFound
		}

		ambulance.WaitingList = append(ambulance.WaitingList[:entryIndx], ambulance.WaitingList[entryIndx+1:]...)
		ambulance.reconcileWaitingList()
		return ambulance, nil, http.StatusNoContent
	})
}

// GetWaitingListEntries - Provides the ambulance waiting list
func (o *implAmbulanceWaitingListAPI) GetWaitingListEntries(c *gin.Context) {
	updateAmbulanceFunc(c, func(c *gin.Context, ambulance *Ambulance) (*Ambulance, interface{}, int) {
		result := ambulance.WaitingList
		if result == nil {
			result = []WaitingListEntry{}
		}
		return nil, result, http.StatusOK
	})
}

// GetWaitingListEntry - Provides details about waiting list entry
func (o *implAmbulanceWaitingListAPI) GetWaitingListEntry(c *gin.Context) {
	updateAmbulanceFunc(c, func(c *gin.Context, ambulance *Ambulance) (*Ambulance, interface{}, int) {
		entryID := c.Param("entryId")
		if entryID == "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Entry ID is required",
			}, http.StatusBadRequest
		}

		entryIndx := slices.IndexFunc(ambulance.WaitingList, func(waiting WaitingListEntry) bool {
			return entryID == waiting.Id
		})
		if entryIndx < 0 {
			return nil, gin.H{
				"status":  http.StatusNotFound,
				"message": "Entry not found",
			}, http.StatusNotFound
		}

		return nil, ambulance.WaitingList[entryIndx], http.StatusOK
	})
}

// UpdateWaitingListEntry - Updates specific entry
func (o *implAmbulanceWaitingListAPI) UpdateWaitingListEntry(c *gin.Context) {
	updateAmbulanceFunc(c, func(c *gin.Context, ambulance *Ambulance) (*Ambulance, interface{}, int) {
		var entry WaitingListEntry
		if err := c.ShouldBindJSON(&entry); err != nil {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request body",
				"error":   err.Error(),
			}, http.StatusBadRequest
		}

		entryID := c.Param("entryId")
		if entryID == "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Entry ID is required",
			}, http.StatusBadRequest
		}

		entryIndx := slices.IndexFunc(ambulance.WaitingList, func(waiting WaitingListEntry) bool {
			return entryID == waiting.Id
		})
		if entryIndx < 0 {
			return nil, gin.H{
				"status":  http.StatusNotFound,
				"message": "Entry not found",
			}, http.StatusNotFound
		}

		if entry.PatientId != "" {
			ambulance.WaitingList[entryIndx].PatientId = entry.PatientId
		}
		if entry.Id != "" {
			ambulance.WaitingList[entryIndx].Id = entry.Id
		}
		if entry.Name != "" {
			ambulance.WaitingList[entryIndx].Name = entry.Name
		}
		if entry.WaitingSince.After(time.Time{}) {
			ambulance.WaitingList[entryIndx].WaitingSince = entry.WaitingSince
		}
		if entry.EstimatedDurationMinutes > 0 {
			ambulance.WaitingList[entryIndx].EstimatedDurationMinutes = entry.EstimatedDurationMinutes
		}
		if entry.Condition.Value != "" {
			ambulance.WaitingList[entryIndx].Condition = entry.Condition
		}

		ambulance.reconcileWaitingList()
		return ambulance, ambulance.WaitingList[entryIndx], http.StatusOK
	})
}
