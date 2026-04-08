package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
	taskdomain "example.com/taskservice/internal/domain/task"
	recurrenceusecase "example.com/taskservice/internal/usecase/recurrence"
)

type RecurrenceHandler struct {
	usecase recurrenceusecase.Usecase
}

func NewRecurrenceHandler(usecase recurrenceusecase.Usecase) *RecurrenceHandler {
	return &RecurrenceHandler{usecase: usecase}
}

// Create создает правило периодичности и генерирует задачи
// POST /api/v1/recurrence-rules
func (h *RecurrenceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRecurrenceRequestDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	// Валидация статуса (если не указан — new)
	if req.Task.Status == "" {
		req.Task.Status = taskdomain.StatusNew
	}

	out, err := h.usecase.Create(r.Context(), recurrenceusecase.CreateInput{
		TaskTitle:       req.Task.Title,
		TaskDescription: req.Task.Description,
		TaskStatus:      req.Task.Status,
		Type:            req.Type,
		Params:          req.Params,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
	})
	if err != nil {
		writeRecurrenceError(w, err)
		return
	}

	// Преобразуем задачи в DTO
	taskDTOs := make([]taskDTO, 0, len(out.Tasks))
	for i := range out.Tasks {
		taskDTOs = append(taskDTOs, newTaskDTO(&out.Tasks[i]))
	}

	writeJSON(w, http.StatusCreated, createRecurrenceResponseDTO{
		Rule:  newRecurrenceRuleDTO(out.Rule),
		Tasks: taskDTOs,
	})
}

// GetByID возвращает правило по ID
// GET /api/v1/recurrence-rules/{id}
func (h *RecurrenceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getRecurrenceID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	rule, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeRecurrenceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newRecurrenceRuleDTO(rule))
}

// List возвращает все правила
// GET /api/v1/recurrence-rules
func (h *RecurrenceHandler) List(w http.ResponseWriter, r *http.Request) {
	rules, err := h.usecase.List(r.Context())
	if err != nil {
		writeRecurrenceError(w, err)
		return
	}

	response := make([]recurrenceRuleDTO, 0, len(rules))
	for i := range rules {
		response = append(response, newRecurrenceRuleDTO(&rules[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

// Delete удаляет правило и все связанные задачи
// DELETE /api/v1/recurrence-rules/{id}
func (h *RecurrenceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getRecurrenceID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		writeRecurrenceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getRecurrenceID извлекает ID из URL
func getRecurrenceID(r *http.Request) (int64, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		return 0, errors.New("missing recurrence rule id")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid recurrence rule id")
	}

	return id, nil
}

// writeRecurrenceError обрабатывает ошибки usecase
func writeRecurrenceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, recurrencedomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, recurrenceusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}
