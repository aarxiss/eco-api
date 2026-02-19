package handler

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"eco-api/internal/blockchain"
	"eco-api/internal/model"
	"eco-api/internal/repository"
)

type MeasurementHandler struct {
	repo     *repository.MeasurementRepo
	bcClient *blockchain.AnchorClient
}

func NewMeasurementHandler(repo *repository.MeasurementRepo, bcClient *blockchain.AnchorClient) *MeasurementHandler {
	return &MeasurementHandler{repo: repo, bcClient: bcClient}
}

// @Summary Read Measurements
// @Tags Measurements
// @Produce json
// @Success 200 {array} model.Measurement
// @Router /measurements [get]
func (h *MeasurementHandler) GetMeasurements(w http.ResponseWriter, r *http.Request) {
	results, err := h.repo.GetAll()
	if err != nil {
		http.Error(w, "Помилка читання з бази даних", http.StatusInternalServerError)
		return
	}
	if results == nil {
		results = []model.Measurement{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// @Summary Create Measurement
// @Tags Measurements
// @Accept json
// @Produce plain
// @Param request body model.Measurement true "Дані вимірювання"
// @Success 201 {string} string "Дані успішно збережено!"
// @Router /measurements [post]
func (h *MeasurementHandler) CreateMeasurement(w http.ResponseWriter, r *http.Request) {
	var m model.Measurement
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "Неправильний формат даних", http.StatusBadRequest)
		return
	}

	if m.SensorID == "" {
		http.Error(w, "Помилка: sensor_id є обов'язковим", http.StatusBadRequest)
		return
	}

	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}

	if err := h.repo.Create(m); err != nil {
		http.Error(w, "Помилка сервера при збереженні", http.StatusInternalServerError)
		return
	}

	if h.bcClient != nil {
		go func(id string, val float64) {
			txHash, err := h.bcClient.AnchorData(context.Background(), id, val)
			if err != nil {
				log.Printf("Помилка блокчейну для сенсора %s: %v", id, err)
			} else {
				log.Printf(" Дані захешовано в блокчейні! Хеш: %s", txHash)
			}
		}(m.SensorID, m.Value)
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Дані успішно збережено!")
}

// @Summary Update Measurement
// @Tags Measurements
// @Accept json
// @Produce plain
// @Param m_id path string true "ID сенсора"
// @Param request body model.Measurement true "Нове значення"
// @Success 200 {string} string "Оновлено"
// @Router /measurements/{m_id} [put]
func (h *MeasurementHandler) UpdateMeasurement(w http.ResponseWriter, r *http.Request) {
	mID := r.PathValue("m_id")
	var m model.Measurement
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "Помилка декодування", http.StatusBadRequest)
		return
	}

	if err := h.repo.Update(mID, m.Value); err != nil {
		http.Error(w, "Не знайдено", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Дані успішно оновлено!")
}

// @Summary Delete Measurement
// @Tags Measurements
// @Produce plain
// @Param m_id path string true "ID сенсора"
// @Success 200 {string} string "Видалено"
// @Router /measurements/{m_id} [delete]
func (h *MeasurementHandler) DeleteMeasurement(w http.ResponseWriter, r *http.Request) {
	mID := r.PathValue("m_id")
	if err := h.repo.Delete(mID); err != nil {
		http.Error(w, "Не знайдено", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Дані успішно видалено!")
}

// @Summary Verify Measurement
// @Tags Measurements
// @Produce json
// @Param m_id path string true "ID сенсора для перевірки останнього запису"
// @Success 200 {object} map[string]interface{}
// @Router /measurements/{m_id}/verify [get]
func (h *MeasurementHandler) VerifyMeasurement(w http.ResponseWriter, r *http.Request) {
	mID := r.PathValue("m_id")

	m, err := h.repo.GetByID(mID)
	if err != nil {
		http.Error(w, "Дані не знайдено в базі", http.StatusNotFound)
		return
	}

	dataString := fmt.Sprintf("%s:%.2f", m.SensorID, m.Value)
	hash := sha256.Sum256([]byte(dataString))

	isAnchored := false
	if h.bcClient != nil {

		var errBc error
		isAnchored, errBc = h.bcClient.IsAnchored(context.Background(), hash)
		if errBc != nil {
			log.Printf("Помилка запиту до блокчейну: %v", errBc)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	status := " ДАНІ ВАЛІДНІ"
	if !isAnchored {
		status = " ДАНІ ПІДРОБЛЕНІ АБО НЕ ЗНАЙДЕНІ"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"sensor_id":  m.SensorID,
		"db_value":   m.Value,
		"timestamp":  m.CreatedAt,
		"is_trusted": isAnchored,
		"verdict":    status,
	})
}
