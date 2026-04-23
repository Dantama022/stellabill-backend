package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"stellarbill-backend/internal/repository"
)

// Plan is the plans payload shape exposed by handlers.
type Plan struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Interval    string `json:"interval"`
	Description string `json:"description,omitempty"`
}

var planRepo repository.PlanRepository

// SetPlanRepository allows wiring a PlanRepository (used by routes.Register).
func SetPlanRepository(r repository.PlanRepository) {
	planRepo = r
}

// ListPlans handles requests through the Handler dependency interface.
func (h *Handler) ListPlans(c *gin.Context) {
	plans, err := h.Plans.ListPlans(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if plans == nil {
		plans = []Plan{}
	}
	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

// ListPlans handles global route registration by using the configured repository.
func ListPlans(c *gin.Context) {
	if planRepo == nil {
		c.JSON(http.StatusOK, gin.H{"plans": []Plan{}})
		return
	}

	rows, err := planRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	out := make([]Plan, 0, len(rows))
	for _, r := range rows {
		out = append(out, Plan{
			ID:          r.ID,
			Name:        r.Name,
			Amount:      r.Amount,
			Currency:    r.Currency,
			Interval:    r.Interval,
			Description: r.Description,
		})
	}
	c.JSON(http.StatusOK, gin.H{"plans": out})
}
