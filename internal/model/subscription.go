package model

import (
	"effective_mobile/internal/logger"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
)

type Subscription struct {
	ID          uuid.UUID  `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   MonthYear  `json:"start_date"`
	EndDate     *MonthYear `json:"end_date,omitempty"`
}
type MonthYear struct {
	time.Time
}

func (my *MonthYear) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	t, err := time.Parse("01-2006", s)
	if err != nil {
		logger.Log.Warnf("Не удалось распарсить дату '%s': %v", s, err)
		return errors.New("дата должна быть в формате MM-YYYY")
	}
	my.Time = t
	return nil
}

func (my MonthYear) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%02d-%04d\"", my.Month(), my.Year())), nil
}
