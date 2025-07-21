package database

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type SubscriptionModel struct {
	DB *pgx.Conn
}

type Subscription struct {
	Id          int    `json:"id"`
	ServiceName string `json:"service_name" binding:"required"`
	Price       int    `json:"price" binding:"required"`
	UserId      int    `json:"user_id" binding:"required"`
	StartDate   string `json:"start_date" binding:"required,datetime=02-2006"`
	EndDate     string `json:"end_date" binding:"datetime=02-2006|len=0"`
}

func parseGetDate(start, end string) (parsedStart *string, parsedEnd *string, err error) {
	endDate := "NULL"
	if end != "" {
		endTime, err := time.Parse("01-2006", end)
		if err != nil {
			slog.Error("ERROR in Subscription parseGetDate", "error", err)
			return nil, nil, err
		}

		endDate = fmt.Sprintf("'%s'", endTime.Format("2006-01-02"))
	}

	startTime, err := time.Parse("01-2006", start)
	if err != nil {
		slog.Error("ERROR in Subscription parseGetDate", "error", err)
		return nil, nil, err
	}
	startDate := fmt.Sprintf("'%s'", startTime.Format("2006-01-02"))

	return &startDate, &endDate, nil
}

func (m *SubscriptionModel) Insert(sub *Subscription) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	startDate, endDate, err := parseGetDate(sub.StartDate, sub.EndDate)
	if err != nil {
		slog.Error("ERROR in Subscription Insert", "error", err)
		return err
	}

	query := fmt.Sprintf("INSERT INTO subscription (service_name, price, user_id, start_date, end_date) VALUES ($1,$2,$3,%s,%s) RETURNING id", *startDate, *endDate)

	return m.DB.QueryRow(ctx, query, sub.ServiceName, sub.Price, sub.UserId).Scan(&sub.Id)
}

func (m *SubscriptionModel) Get(id int) (*Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT * FROM subscription WHERE id = $1"

	var sub Subscription
	var startTime time.Time
	var endTime any

	err := m.DB.QueryRow(ctx, query, id).Scan(&sub.Id, &sub.ServiceName, &sub.Price, &sub.UserId, &startTime, &endTime)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		slog.Error("ERROR in Subscription Get", "error", err)
		return nil, err
	}

	sub.StartDate = startTime.Format("01-2006")

	t, ok := endTime.(time.Time)
	if ok {
		sub.EndDate = t.Format("01-2006")
	}

	return &sub, nil
}

func (m *SubscriptionModel) Update(sub *Subscription) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	startDate, endDate, err := parseGetDate(sub.StartDate, sub.EndDate)
	if err != nil {
		slog.Error("ERROR in Subscription Update", "error", err)
		return err
	}

	query := fmt.Sprintf("UPDATE subscription SET service_name = $1, price = $2, user_id = $3, start_date = %s, end_date = %s WHERE id = $4", *startDate, *endDate)

	_, err = m.DB.Exec(ctx, query, sub.ServiceName, sub.Price, sub.UserId, sub.Id)
	if err != nil {
		slog.Error("ERROR in Subscription Update", "error", err)
		return err
	}

	return nil
}

func (m *SubscriptionModel) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "DELETE FROM subscription WHERE id = $1"

	_, err := m.DB.Exec(ctx, query, id)
	if err != nil {
		slog.Error("ERROR in Subscription Delete", "error", err)
		return err
	}

	return nil
}

func (m *SubscriptionModel) GetList(filter map[string]string) ([]*Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var filterSB strings.Builder
	for k, v := range filter {
		var str string
		if filterSB.Len() == 0 {
			str = fmt.Sprintf("WHERE %s = '%s'", k, v)
		} else {
			str = fmt.Sprintf("AND %s = '%s'", k, v)
		}
		filterSB.WriteString(str)
	}

	query := fmt.Sprintf("SELECT * FROM subscription %s", filterSB.String())

	rows, err := m.DB.Query(ctx, query)
	if err != nil {
		slog.Error("ERROR in Subscription GetList", "error", err)
		return nil, err
	}

	defer rows.Close()

	subs := []*Subscription{}

	for rows.Next() {
		var sub Subscription

		var startTime time.Time
		var endTime any

		err := rows.Scan(&sub.Id, &sub.ServiceName, &sub.Price, &sub.UserId, &startTime, &endTime)
		if err != nil {
			slog.Error("ERROR in Subscription GetList", "error", err)
			return nil, err
		}
		sub.StartDate = startTime.Format("01-2006")

		t, ok := endTime.(time.Time)
		if ok {
			sub.EndDate = t.Format("01-2006")
		}

		subs = append(subs, &sub)
	}

	if err = rows.Err(); err != nil {
		slog.Error("ERROR in Subscription GetList", "error", err)
		return nil, err
	}

	return subs, nil
}

func (m *SubscriptionModel) GetPrice(startPeriodInput, endPeriodInput time.Time, filter map[string]string) (totalPrice int, prices map[int]string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	prices = make(map[int]string)
	var filterSB strings.Builder
	for k, v := range filter {
		str := fmt.Sprintf(" AND %s = '%s'", k, v)
		filterSB.WriteString(str)
	}

	query := fmt.Sprintf(`SELECT *
	 		FROM subscription
			WHERE start_date <= '%s'
			AND (end_date >= '%s'
			OR end_date IS NULL)
			 %s`, endPeriodInput.Format("2006/01/02"),
		startPeriodInput.Format("2006/01/02"),
		filterSB.String())

	// fmt.Println(query)

	rows, err := m.DB.Query(ctx, query)
	if err != nil {
		slog.Error("ERROR in Subscription GetPrice", "error", err)
		return totalPrice, prices, err
	}

	for rows.Next() {
		var sub Subscription

		var startSub time.Time
		var endAny any
		endSub := endPeriodInput
		err := rows.Scan(&sub.Id, &sub.ServiceName, &sub.Price, &sub.UserId, &startSub, &endAny)
		if err != nil {
			slog.Error("ERROR in Subscription GetPrice", "error", err)
			return totalPrice, prices, err
		}

		t, ok := endAny.(time.Time)
		if ok {
			endSub = t
		}

		startPeriod := startPeriodInput
		if startSub.After(startPeriod) {
			startPeriod = startSub
		}

		endPeriod := endPeriodInput
		if endSub.Before(endPeriod) {
			endPeriod = endSub
		}

		y1, m1, _ := endPeriod.Date()
		y2, m2, _ := startPeriod.Date()

		months := (y1-y2)*12 + int(m1) - int(m2)

		prices[sub.Id] = fmt.Sprintf("service_name: %s, months: %d, price: %d, user_id: %d, total_price: %d", sub.ServiceName, months, sub.Price, sub.UserId, months*sub.Price)
		totalPrice += months * sub.Price
	}

	defer rows.Close()

	return totalPrice, prices, nil
}
