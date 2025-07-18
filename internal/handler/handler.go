package handler

import (
	"database/sql"
	"effective_mobile/internal/logger"
	"effective_mobile/internal/model"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type Handler struct {
	DB *sql.DB
}

func New(db *sql.DB) *Handler {
	return &Handler{DB: db}
}

// CreateSubscription godoc
// @Summary Создать подписку
// @Description  Создаёт новую подписку для пользователя
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body model.Subscription true "Данные подписки"
// @Success 201 {object} map[string]string
// @Failure 400 {string} string "Неверный JSON"
// @Failure 500 {string} string "Ошибка базы данных"
// @Router /subscriptions [post]
func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("CreateSubscription: получен POST-запрос")
	var sub model.Subscription

	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, "Невалидный JSON", http.StatusBadRequest)
		logger.Log.Warnf("Ошибка декодирования: %v", err)
		return
	}

	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}

	query := `INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := h.DB.Exec(
		query,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate.Time,
		nilIfEmpty(sub.EndDate),
	)
	if err != nil {
		http.Error(w, "Ошибка вставки в базу", http.StatusInternalServerError)
		logger.Log.Errorf("Ошибка вставки: %v", err)
		return
	}

	logger.Log.Infof("Подписка успешно создана: ID=%s", sub.ID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": sub.ID,
	})
}

func nilIfEmpty(t *model.MonthYear) interface{} {
	if t == nil || t.IsZero() {
		return nil
	}
	return t.Time
}

// GetAllSubscriptions godoc
// @Summary Получить список всех подписок
// @Description  Возвращает все подписки из базы данных
// @Tags subscriptions
// @Produce json
// @Success 200 {array} model.Subscription
// @Failure 500 {string} {string} string "Ошибка базы данных"
// @Router /subscriptions [get]
func (h *Handler) GetAllSubscriptions(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("Запрос на получение всех подписок")
	rows, err := h.DB.Query(`SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions`)
	if err != nil {
		http.Error(w, "Ошибка получения данных", http.StatusInternalServerError)
		logger.Log.Errorf("Ошибка выборки: %v", err)
		return
	}
	defer rows.Close()

	var subscriptions []model.Subscription

	for rows.Next() {
		var sub model.Subscription
		var endDate sql.NullTime

		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate.Time,
			&endDate,
		)
		if err != nil {
			logger.Log.Errorf("Ошибка сканирования строки: %v", err)
			continue
		}
		if endDate.Valid {
			sub.EndDate = &model.MonthYear{Time: endDate.Time}
		}
		subscriptions = append(subscriptions, sub)
	}

	logger.Log.Infof("Получено %d подписок", len(subscriptions))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscriptions)
}

// GetTotalPrice godoc
// @Summary Получить суммарную стоимость подписок с фильтрацией
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "UUID пользователя"
// @Param service_name query string false "Название сервиса"
// @Param from query string false "Начало периода (MM-YYYY)"
// @Param to query string false "Конец периода (MM-YYYY)"
// @Success 200 {object} map[string]int64
// @Failure 400 {string} string "Ошибка формата даты или UUID"
// @Failure 500 {string} string "Ошибка базы данных"
// @Router /subscriptions/total [get]
func (h *Handler) GetTotalPrice(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("Получен запрос на подсчёт суммы подписок")
	q := r.URL.Query()

	userID := q.Get("user_id")
	serviceName := q.Get("service_name")
	fromStr := q.Get("from")
	toStr := q.Get("to")

	sqlStr := `SELECT SUM(price) FROM subscriptions WHERE 1=1`
	args := []interface{}{}
	argID := 1

	if userID != "" {
		sqlStr += fmt.Sprintf(" AND user_id = $%d", argID)
		args = append(args, userID)
		argID++
	}

	if serviceName != "" {
		sqlStr += fmt.Sprintf(" AND service_name = $%d", argID)
		args = append(args, serviceName)
		argID++
	}

	if fromStr != "" {
		fromDate, err := time.Parse("01-2006", fromStr)
		if err != nil {
			http.Error(w, "Неверный формат from (MM-YYYY)", http.StatusBadRequest)
			logger.Log.Warnf("Неверный формат from: %s", fromStr)
			return
		}
		sqlStr += fmt.Sprintf(" AND start_date >= $%d", argID)
		args = append(args, fromDate)
		argID++
	}

	if toStr != "" {
		toDate, err := time.Parse("01-2006", toStr)
		if err != nil {
			http.Error(w, "Неверный формат to (MM-YYYY)", http.StatusBadRequest)
			logger.Log.Warnf("Неверный формат to: %s", toStr)
			return
		}

		toDate = toDate.AddDate(0, 1, -1)
		sqlStr += fmt.Sprintf(" AND start_date <= $%d", argID)
		args = append(args, toDate)
		argID++
	}

	var total sql.NullInt64
	err := h.DB.QueryRow(sqlStr, args...).Scan(&total)
	if err != nil {
		http.Error(w, "Ошибка запроса", http.StatusInternalServerError)
		logger.Log.Errorf("Ошибка подсчёта: %v", err)
		return
	}

	logger.Log.Infof("Общая сумма подписок: %d", total.Int64)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{
		"total_price": total.Int64,
	})
}

// GetSubscriptionByID godoc
// @Summary Получить подписку по ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки (UUID)"
// @Success 200 {object} model.Subscription
// @Failure 400 {string} string "Неверный UUID"
// @Failure 404 {string} string "Подписка не найдена"
// @Failure 500 {string} string "Ошибка базы данных"
// @Router /subscriptions/{id} [get]
func (h *Handler) GetSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("Получен запрос на получение подписки по ID")
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Неверный UUID", http.StatusBadRequest)
		logger.Log.Warnf("Невалидный UUID: %v", idStr)
		return
	}

	var sub model.Subscription
	var endDate sql.NullTime

	query := `SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions
		WHERE id = $1`

	err = h.DB.QueryRow(query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate.Time,
		&endDate,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "Подписка не найдена", http.StatusNotFound)
		logger.Log.Infof("Подписка с ID %s не найдена", id)
		return
	} else if err != nil {
		http.Error(w, "Ошибка получения подписки", http.StatusInternalServerError)
		logger.Log.Errorf("Ошибка запроса по id: %v", err)
		return
	}

	if endDate.Valid {
		sub.EndDate = &model.MonthYear{Time: endDate.Time}
	}

	logger.Log.Infof("Подписка с ID %s найдена", id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}

// UpdateSubscription godoc
// @Summary Обновить подписку по ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки (UUID)"
// @Param subscription body model.Subscription true "Новые данные"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Невалидный UUID или JSON"
// @Failure 404 {string} string "Подписка не найдена"
// @Failure 500 {string} string "Ошибка базы данных"
// @Router /subscriptions/{id} [put]
func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("Получен запрос на обновление подписки")
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Неверный UUID", http.StatusBadRequest)
		logger.Log.Warnf("Невалидный UUID: %v", idStr)
		return
	}

	var sub model.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, "Невалидный JSON", http.StatusBadRequest)
		logger.Log.Warnf("Ошибка парсинга тела: %v", err)
		return
	}

	query := `UPDATE subscriptions SET service_name = $1,
			price = $2,
			user_id = $3,
			start_date = $4,
			end_date = $5
		WHERE id = $6`

	res, err := h.DB.Exec(
		query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate.Time,
		nilIfEmpty(sub.EndDate),
		id,
	)
	if err != nil {
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
		logger.Log.Errorf("Ошибка обновления по id %s: %v", id, err)
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		http.Error(w, "Подписка не найдена", http.StatusNotFound)
		logger.Log.Warnf("Подписка с id %s не найдена для обновления", id)
		return
	}

	logger.Log.Infof("Подписка с id %s успешно обновлена", id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Подписка обновлена",
	})
}

// DeleteSubscription godoc
// @Summary Удалить подписку по ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки (UUID)"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Невалидный UUID"
// @Failure 404 {string} string "Подписка не найдена"
// @Failure 500 {string} string "Ошибка базы данных"
// @Router /subscriptions/{id} [delete]
func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	logger.Log.Info("Получен запрос на удаление подписки")
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Неверный UUID", http.StatusBadRequest)
		logger.Log.Warnf("Невалидный UUID для удаления: %v", idStr)
		return
	}

	query := `DELETE FROM subscriptions WHERE id = $1`

	res, err := h.DB.Exec(query, id)
	if err != nil {
		http.Error(w, "Ошибка удаления", http.StatusInternalServerError)
		logger.Log.Errorf("Ошибка удаления %s: %v", id, err)
		return
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		http.Error(w, "Подписка не найдена", http.StatusNotFound)
		logger.Log.Warnf("Подписка с id %s не найдена для удаления", id)
		return
	}

	logger.Log.Infof("Подписка с id %s успешно удалена", id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Подписка удалена",
	})
}
