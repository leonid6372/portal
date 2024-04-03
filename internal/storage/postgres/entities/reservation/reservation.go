package reservation

import (
	"fmt"
	"portal/internal/storage/postgres"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	qrGetPlaceById        = `SELECT name, properties FROM place WHERE place_id = $1`
	qrReservationInsert   = `INSERT INTO reservation (place_id, start, finish, user_id) VALUES ($1, $2, $3, 1)` // сделать $4
	qrGetActualPlaceList  = `SELECT jsonb_agg(actual_places) FROM actual_places;`
	qrGetUserReservations = `SELECT reservation_id, place_id, start, finish FROM reservation WHERE user_id = $1`
	qrReservationUpdate   = `UPDATE reservation SET place_id = $2, start = $3, finish = $4 WHERE reservation_id = $1`
	qrReservationDrop     = `DELETE FROM reservation WHERE reservation_id = $1`
)

type Place struct {
	PlaceID    int    `json:"place_id"`
	Name       string `json:"name"`
	Properties string `json:"properties"`
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
	ReservationID int              `json:"reservation_id"`
	PlaceID       int              `json:"place_id"`
	Start         pgtype.Timestamp `json:"start"`
	Finish        pgtype.Timestamp `json:"finish"`
	UserID        int              `json:"user_id"`
}

func (r *Reservation) ReservationInsert(storage *postgres.Storage, placeID int, start, finish string) error {
	const op = "storage.postgres.entities.reservation.ReservationInsert" // Имя текущей функции для логов и ошибок

	_, err := storage.DB.Exec(qrReservationInsert, placeID, start, finish) // добавить userID
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Reservation) GetUserReservations(storage *postgres.Storage, userID int) (string, error) {
	const op = "storage.postgres.entities.reservation.GetUserReservations" // Имя текущей функции для логов и ошибок

	qrResult, err := storage.DB.Query(qrGetUserReservations, userID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var userReservations string
	for qrResult.Next() {
		if err := qrResult.Scan(&userReservations); err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}
	}

	return userReservations, nil
}

func (r *Reservation) ReservationUpdate(storage *postgres.Storage, reservationID, placeID int, start, finish pgtype.Timestamp) error {
	const op = "storage.postgres.entities.reservation.ReservationUpdate" // Имя текущей функции для логов и ошибок

	_, err := storage.DB.Exec(qrReservationUpdate, reservationID, placeID, start, finish)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Reservation) ReservationDrop(storage *postgres.Storage, reservationID int) error {
	const op = "storage.postgres.entities.reservation.ReservationDrop" // Имя текущей функции для логов и ошибок

	_, err := storage.DB.Exec(qrReservationDrop, reservationID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
