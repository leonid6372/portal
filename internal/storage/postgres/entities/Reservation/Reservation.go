package Reservation

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	db "portal/internal/storage/postgres"
)

type Place struct {
	PlaceID    int
	Name       string
	Properties string
	IsAvalible bool
}

type Reservation struct {
	ReservationID int
	PlaceID       int
	Start         pgtype.Timestamp
	End           pgtype.Timestamp
	UserID        int
}

const (
	qrGetPlaceById      = `SELECT name, properties FROM place WHERE place_id = $1`
	qrReservationInsert = `INSERT INTO reservation (place_id, start, end, user_id) VALUES ($1, $2, $3, 1)`

	qrGetActualPlaceList = `SELECT jsonb_agg(ActualPlaces) FROM ActualPlaces;`
)

func (p *Place) GetPlaceById(db *db.Storage) (bool, error) {
	const op = "storage.postgres.entities.getPlaceById" // Имя текущей функции для логов и ошибок
	qrResult, err := db.DB.Query(qrGetPlaceById, p.PlaceID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	for qrResult.Next() {
		if err := qrResult.Scan(&p.Name, &p.Properties); err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
	}
	return true, nil
}

func (r *Reservation) ReservationInsert(db *db.Storage, placeID int, start, finish string) (bool, error) {
	const op = "storage.postgres.entities.reservationInsert" // Имя текущей функции для логов и ошибок
	_, err := db.DB.Query(qrReservationInsert, placeID, start, finish)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return true, nil
}

func (p *Place) GetActualPlaceList(db *db.Storage) (string, error) {
	const op = "storage.postgres.GetActualPlaceList" // Имя текущей функции для логов и ошибок

	qrResult, err := db.DB.Query(qrGetActualPlaceList)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var placeList string
	for qrResult.Next() {
		if err := qrResult.Scan(&placeList); err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}
	}

	return placeList, nil
}
