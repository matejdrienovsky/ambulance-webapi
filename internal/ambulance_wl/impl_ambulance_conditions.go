package ambulance_wl

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

type implAmbulanceConditionsAPI struct{}

func NewAmbulanceConditionsApi() AmbulanceConditionsAPI {
	return &implAmbulanceConditionsAPI{}
}

// GetConditions - Provides the list of conditions associated with ambulance
func (o *implAmbulanceConditionsAPI) GetConditions(c *gin.Context) {
	updateAmbulanceFunc(c, func(
		c *gin.Context,
		ambulance *Ambulance,
	) (updatedAmbulance *Ambulance, responseContent interface{}, status int) {
		result := ambulance.PredefinedConditions
		if result == nil {
			result = []Condition{}
		}
		return nil, result, http.StatusOK
	})
}

// CreateCondition - Adds new predefined condition
func (o *implAmbulanceConditionsAPI) CreateCondition(c *gin.Context) {
	updateAmbulanceFunc(c, func(
		c *gin.Context,
		ambulance *Ambulance,
	) (updatedAmbulance *Ambulance, responseContent interface{}, status int) {
		var condition Condition
		if err := c.ShouldBindJSON(&condition); err != nil {
			return nil, gin.H{
				"status":  "Bad Request",
				"message": "Invalid request body",
				"error":   err.Error(),
			}, http.StatusBadRequest
		}

		condition.Code = strings.TrimSpace(condition.Code)
		condition.Value = strings.TrimSpace(condition.Value)
		if condition.Code == "" || condition.Value == "" {
			return nil, gin.H{
				"status":  "Bad Request",
				"message": "condition code and value are required",
			}, http.StatusBadRequest
		}

		conflictIdx := slices.IndexFunc(ambulance.PredefinedConditions, func(existing Condition) bool {
			return strings.EqualFold(existing.Code, condition.Code)
		})
		if conflictIdx >= 0 {
			return nil, gin.H{
				"status":  "Conflict",
				"message": "Condition already exists",
			}, http.StatusConflict
		}

		ambulance.PredefinedConditions = append(ambulance.PredefinedConditions, condition)
		return ambulance, condition, http.StatusCreated
	})
}

// DeleteCondition - Deletes one predefined condition
func (o *implAmbulanceConditionsAPI) DeleteCondition(c *gin.Context) {
	updateAmbulanceFunc(c, func(
		c *gin.Context,
		ambulance *Ambulance,
	) (updatedAmbulance *Ambulance, responseContent interface{}, status int) {
		conditionCode := strings.TrimSpace(c.Param("conditionCode"))
		if conditionCode == "" {
			return nil, gin.H{
				"status":  "Bad Request",
				"message": "conditionCode is required",
			}, http.StatusBadRequest
		}

		index := slices.IndexFunc(ambulance.PredefinedConditions, func(existing Condition) bool {
			return strings.EqualFold(existing.Code, conditionCode)
		})
		if index < 0 {
			return nil, gin.H{
				"status":  "Not Found",
				"message": "Condition not found",
			}, http.StatusNotFound
		}

		ambulance.PredefinedConditions = append(
			ambulance.PredefinedConditions[:index],
			ambulance.PredefinedConditions[index+1:]...,
		)
		return ambulance, nil, http.StatusNoContent
	})
}

// UpdateConditions - Replaces predefined conditions list
func (o *implAmbulanceConditionsAPI) UpdateConditions(c *gin.Context) {
	updateAmbulanceFunc(c, func(
		c *gin.Context,
		ambulance *Ambulance,
	) (updatedAmbulance *Ambulance, responseContent interface{}, status int) {
		conditions := []Condition{}
		if err := c.ShouldBindJSON(&conditions); err != nil {
			return nil, gin.H{
				"status":  "Bad Request",
				"message": "Invalid request body",
				"error":   err.Error(),
			}, http.StatusBadRequest
		}

		seenCodes := map[string]struct{}{}
		for i := range conditions {
			conditions[i].Code = strings.TrimSpace(conditions[i].Code)
			conditions[i].Value = strings.TrimSpace(conditions[i].Value)

			if conditions[i].Code == "" || conditions[i].Value == "" {
				return nil, gin.H{
					"status":  "Bad Request",
					"message": "each condition must contain code and value",
				}, http.StatusBadRequest
			}

			key := strings.ToLower(conditions[i].Code)
			if _, exists := seenCodes[key]; exists {
				return nil, gin.H{
					"status":  "Conflict",
					"message": "Duplicate condition code in request",
				}, http.StatusConflict
			}
			seenCodes[key] = struct{}{}
		}

		ambulance.PredefinedConditions = conditions
		return ambulance, conditions, http.StatusOK
	})
}
