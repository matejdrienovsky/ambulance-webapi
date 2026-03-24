package ambulance_wl

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/matejdrienovsky/ambulance-webapi/internal/db_service"
)

type implAmbulancesAPI struct{}

func NewAmbulancesApi() AmbulancesAPI {
	return &implAmbulancesAPI{}
}

// CreateAmbulance - Saves new ambulance definition
func (o *implAmbulancesAPI) CreateAmbulance(c *gin.Context) {
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db not found",
				"error":   "db not found",
			},
		)
		return
	}

	db, ok := value.(db_service.DbService[Ambulance])
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db context is not of required type",
				"error":   "cannot cast db context to db_service.DbService",
			},
		)
		return
	}

	ambulance := Ambulance{}
	if err := c.ShouldBindJSON(&ambulance); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  "Bad Request",
				"message": "Invalid request body",
				"error":   err.Error(),
			},
		)
		return
	}

	if ambulance.Id == "" {
		ambulance.Id = uuid.New().String()
	}

	err := db.CreateDocument(c, ambulance.Id, &ambulance)
	switch err {
	case nil:
		c.JSON(http.StatusCreated, ambulance)
	case db_service.ErrConflict:
		c.JSON(
			http.StatusConflict,
			gin.H{
				"status":  "Conflict",
				"message": "Ambulance already exists",
				"error":   err.Error(),
			},
		)
	default:
		c.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to create ambulance in database",
				"error":   err.Error(),
			},
		)
	}
}

// DeleteAmbulance - Deletes specific ambulance
func (o *implAmbulancesAPI) DeleteAmbulance(c *gin.Context) {
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service not found",
				"error":   "db_service not found",
			},
		)
		return
	}

	db, ok := value.(db_service.DbService[Ambulance])
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service context is not of type db_service.DbService",
				"error":   "cannot cast db_service context to db_service.DbService",
			},
		)
		return
	}

	ambulanceId := c.Param("ambulanceId")
	err := db.DeleteDocument(c, ambulanceId)

	switch err {
	case nil:
		c.AbortWithStatus(http.StatusNoContent)
	case db_service.ErrNotFound:
		c.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Ambulance not found",
				"error":   err.Error(),
			},
		)
	default:
		c.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to delete ambulance from database",
				"error":   err.Error(),
			},
		)
	}
}

// GetAmbulance - Provides details about ambulance
func (o *implAmbulancesAPI) GetAmbulance(c *gin.Context) {
	updateAmbulanceFunc(c, func(
		c *gin.Context,
		ambulance *Ambulance,
	) (updatedAmbulance *Ambulance, responseContent interface{}, status int) {
		return nil, ambulance, http.StatusOK
	})
}

// UpdateAmbulance - Updates specific ambulance
func (o *implAmbulancesAPI) UpdateAmbulance(c *gin.Context) {
	updateAmbulanceFunc(c, func(
		c *gin.Context,
		_ *Ambulance,
	) (updatedAmbulance *Ambulance, responseContent interface{}, status int) {
		ambulanceID := c.Param("ambulanceId")

		var updated Ambulance
		if err := c.ShouldBindJSON(&updated); err != nil {
			return nil, gin.H{
				"status":  "Bad Request",
				"message": "Invalid request body",
				"error":   err.Error(),
			}, http.StatusBadRequest
		}

		if updated.Id == "" {
			updated.Id = ambulanceID
		}
		if updated.Id != ambulanceID {
			return nil, gin.H{
				"status":  "Forbidden",
				"message": "Path ambulanceId and ambulance.id are mismatching",
			}, http.StatusForbidden
		}
		if updated.Name == "" || updated.RoomNumber == "" {
			return nil, gin.H{
				"status":  "Bad Request",
				"message": "name and roomNumber are required",
			}, http.StatusBadRequest
		}

		if len(updated.WaitingList) > 0 {
			updated.reconcileWaitingList()
		}
		return &updated, updated, http.StatusOK
	})
}
