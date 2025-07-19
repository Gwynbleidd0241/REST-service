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
	ID          uuid.UUID  `json:"id" example:"f4e2c8ad-d773-4d39-85b1-a63d2b2cb41f"`
	ServiceName string     `json:"service_name"  example:"Spotify"`
	Price       int        `json:"price" example:"199"`
	UserID      uuid.UUID  `json:"user_id" example:"00000000-0000-0000-0000-000000000001"`
	StartDate   MonthYear  `json:"start_date" swaggertype:"string" example:"07-2025"`
	EndDate     *MonthYear `json:"end_date,omitempty" swaggertype:"string" example:"12-2025"`
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
