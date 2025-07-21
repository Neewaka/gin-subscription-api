package database

import "github.com/jackc/pgx/v5"

type Models struct {
	Subscriptions SubscriptionModel
}

func NewModels(db *pgx.Conn) Models {
	return Models{
		Subscriptions: SubscriptionModel{DB: db},
	}
}
