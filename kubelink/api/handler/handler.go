package handler

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

type Handler interface {
	GetSampleRepsonse(w http.ResponseWriter, r *http.Request)
}

type HandlerImpl struct {
	logger *zap.SugaredLogger
}

func NewHandlerImpl(logger *zap.SugaredLogger) *HandlerImpl {
	return &HandlerImpl{logger: logger}
}

func (impl *HandlerImpl) GetSampleRepsonse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"status":  "success",
		"message": "Sample response",
		"data": map[string]interface{}{
			"id":   1,
			"name": "Demo Object",
		},
	}
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		impl.logger.Errorw("Error in encoding response", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	return
}
