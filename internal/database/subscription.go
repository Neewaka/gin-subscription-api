package database

import (
	"context"
	"fmt"
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
			return nil, nil, err
		}

		endDate = fmt.Sprintf("'%s'", endTime.Format("2006-01-02"))
	}

	startTime, err := time.Parse("01-2006", start)
	if err != nil {
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
		return err
	}

	query := fmt.Sprintf("UPDATE subscription SET service_name = $1, price = $2, user_id = $3, start_date = %s, end_date = %s WHERE id = $4", *startDate, *endDate)

	_, err = m.DB.Exec(ctx, query, sub.ServiceName, sub.Price, sub.UserId, sub.Id)
	if err != nil {
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
		return err
	}

	return nil
}

func (m *SubscriptionModel) GetList() ([]*Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT * FROM subscription"

	rows, err := m.DB.Query(ctx, query)
	if err != nil {
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
		return nil, err
	}

	return subs, nil
}

func (m *SubscriptionModel) GetPrice(startPeriodInput, endPeriodInput time.Time, filter map[string]string) (price int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

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
		return 0, err
	}

	for rows.Next() {
		var sub Subscription

		var startSub time.Time
		var endAny any
		endSub := endPeriodInput
		err := rows.Scan(&sub.Id, &sub.ServiceName, &sub.Price, &sub.UserId, &startSub, &endAny)
		if err != nil {
			return 0, err
		}

		t, ok := endAny.(time.Time)
		if ok {
			endSub = t
		}

		// fmt.Println("period before update")
		// fmt.Println(startPeriodInput, endPeriodInput)
		// fmt.Println("time before update")
		// fmt.Println(startSub, endSub)

		startPeriod := startPeriodInput
		if startSub.After(startPeriod) {
			startPeriod = startSub
		}

		endPeriod := endPeriodInput
		if endSub.Before(endPeriod) {
			endPeriod = endSub
		}

		// fmt.Println("period after update")
		// fmt.Println(startPeriod, endPeriod)

		y1, m1, _ := endPeriod.Date()
		y2, m2, _ := startPeriod.Date()
		// fmt.Printf("\n y1 %d, m1 %d, y1 %d, m2 %d", y1, int(m1), y2, int(m2))

		months := (y1-y2)*12 + int(m1) - int(m2)

		// fmt.Printf("\n months is %d, id %d, price %d", months, sub.Id, sub.Price)

		// fmt.Printf("\n ID %d price change %d, months %d", sub.Id, months*sub.Price, months)
		price += months * sub.Price
	}

	defer rows.Close()

	return price, nil
}
