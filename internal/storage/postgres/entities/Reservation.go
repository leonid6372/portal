package entities

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	db "portal/internal/storage/postgres"
)

type Place struct {
	placeId    int
	name       string
	properties string
	isAvalible bool
}

type Reservation struct {
	reservationId int
	placeId       int
	start         pgtype.Timestamp
	end           pgtype.Timestamp
	userId        int
}

const (
	qrGetPlaceById      = `SELECT name, properties FROM place WHERE place_id = $1`
	qrReservationInsert = `INSERT INTO reservation (place_id, start, end, user_id) VALUES ($1, $2, $3, 1)`

	qrGetActualPlaceList = `SELECT jsonb_agg(ActualPlaces) FROM ActualPlaces;`
)

func (p *Place) GetPlaceById(db *db.Storage) (bool, error) {
	const op = "storage.postgres.entities.getPlaceById" // Имя текущей функции для логов и ошибок
	qrResult, err := db.Db.Query(qrGetPlaceById, p.placeId)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	for qrResult.Next() {
		if err := qrResult.Scan(&p.name, &p.properties); err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
	}
	return true, nil
}

func (r *Reservation) ReservationInsert(db *db.Storage, PlaceId int, Start string, End string) (bool, error) {
	const op = "storage.postgres.entities.reservationInsert" // Имя текущей функции для логов и ошибок
	_, err := db.Db.Query(qrReservationInsert, PlaceId, Start, End)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return true, nil
}

func (p *Place) GetActualPlaceList(db *db.Storage) (string, error) {
	const op = "storage.postgres.GetActualPlaceList" // Имя текущей функции для логов и ошибок

	qrResult, err := db.Db.Query(qrGetActualPlaceList)
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
