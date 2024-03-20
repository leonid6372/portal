package reservation

import (
	"fmt"
	"portal/internal/storage/postgres"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	qrGetPlaceById       = `SELECT name, properties FROM place WHERE place_id = $1`
	qrReservationInsert  = `INSERT INTO reservation (place_id, start, finish, user_id) VALUES ($1, $2, $3, 1)` // сделать $4
	qrGetActualPlaceList = `SELECT jsonb_agg(actual_places) FROM actual_places;`
)

type Place struct {
	PlaceID    int
	Name       string
	Properties string
	IsAvalible bool
}

func (p *Place) GetPlaceById(storage *postgres.Storage) (bool, error) {
	const op = "storage.postgres.entities.reservation.getPlaceById" // Имя текущей функции для логов и ошибок

	qrResult, err := storage.DB.Query(qrGetPlaceById, p.PlaceID)
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

func (p *Place) GetActualPlaceList(storage *postgres.Storage) (string, error) {
	const op = "storage.postgres.entities.reservation.GetActualPlaceList" // Имя текущей функции для логов и ошибок

	qrResult, err := storage.DB.Query(qrGetActualPlaceList)
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

type Reservation struct {
	ReservationID int
	PlaceID       int
	Start         pgtype.Timestamp
	End           pgtype.Timestamp
	UserID        int
}

func (r *Reservation) ReservationInsert(storage *postgres.Storage, placeID int, start, finish string) (bool, error) {
	const op = "storage.postgres.entities.reservation.reservationInsert" // Имя текущей функции для логов и ошибок

	_, err := storage.DB.Exec(qrReservationInsert, placeID, start, finish) // добавить userID
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}
