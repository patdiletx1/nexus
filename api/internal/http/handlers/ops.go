package handlers

import (
	"net/http"

	"nexus/api/internal/observability"
)

type OpsHandler struct {
	Metrics *observability.Metrics
}

func (h OpsHandler) Alerts(w http.ResponseWriter, _ *http.Request) {
	alerts := h.Metrics.EvaluateAlerts(observability.AlertThresholds{})
	writeJSON(w, http.StatusOK, map[string]any{
		"alerts": alerts,
	})
}
