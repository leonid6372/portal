package reservation

import (
	"fmt"
	"portal/internal/storage/postgres"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	// Получение актуальных мест 1. делаем все места "доступно" и вычитаем занятые 2. прибавляем занятые с пометкой "недостпуно"
	qrGetActualPlaces = `(SELECT place_id, "name", true AS is_available FROM place
						  EXCEPT
						  SELECT DISTINCT place_id, "name", true AS is_available FROM place_and_reservation
						  WHERE ($1, $2) OVERLAPS ("start", finish))
						  UNION
						  (SELECT DISTINCT place_id, "name", false AS is_available FROM place_and_reservation
						  WHERE ($1, $2) OVERLAPS ("start", finish));`
	qrGetReservationsByUserID = `SELECT reservation_id, place_id, start, finish FROM reservation WHERE user_id = $1;`
	qrGetIsPlaceAvailable     = `SELECT reservation_id FROM reservation WHERE place_id = $1 AND (start, finish) OVERLAPS ($2, $3);`
	qrInsertReservation       = `INSERT INTO reservation (place_id, start, finish, user_id) VALUES ($1, $3, $4, $2);`
	qrUpdateReservation       = `UPDATE reservation SET place_id = $2, start = $3, finish = $4 WHERE reservation_id = $1;`
	qrDeleteReservation       = `DELETE FROM reservation WHERE reservation_id = $1;`
)

type Place struct {
	PlaceID    int    `json:"place_id,omitempty"`
	Name       string `json:"name,omitempty"`
	Properties string `json:"properties,omitempty"`
}

type ActualPlace struct {
	Place
	IsAvailable bool `json:"is_available,omitempty"`
}

func (ap *ActualPlace) GetActualPlaces(storage *postgres.Storage, properties string, start, finish time.Time) ([]ActualPlace, error) {
	const op = "storage.postgres.entities.reservation.GetActualPlaces"

	qrResult, err := storage.DB.Query(qrGetActualPlaces, start, finish)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var aps []ActualPlace

	for qrResult.Next() {
		var ap ActualPlace
		if err := qrResult.Scan(&ap.PlaceID, &ap.Name, &ap.IsAvailable); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		aps = append(aps, ap)
	}

	return aps, nil
}

type Reservation struct {
	ReservationID int              `json:"reservation_id,omitempty"`
	PlaceID       int              `json:"place_id,omitempty"`
	Start         pgtype.Timestamp `json:"start,omitempty"`
	Finish        pgtype.Timestamp `json:"finish,omitempty"`
	UserID        int              `json:"user_id,omitempty"`
}

func (r *Reservation) InsertReservation(storage *postgres.Storage, placeID, userID int, start, finish string) error {
	const op = "storage.postgres.entities.reservation.InsertReservation" // Имя текущей функции для логов и ошибок

	qrResult, err := storage.DB.Query(qrGetIsPlaceAvailable, placeID, start, finish)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Проверка на пустой ответ
	if qrResult.Next() {
		return fmt.Errorf("%s: place is already taken", op)
	}

	_, err = storage.DB.Exec(qrInsertReservation, placeID, userID, start, finish)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Reservation) UpdateReservation(storage *postgres.Storage, reservationID, placeID int, start, finish time.Time) error {
	const op = "storage.postgres.entities.reservation.UpdateReservation" // Имя текущей функции для логов и ошибок

	_, err := storage.DB.Exec(qrUpdateReservation, reservationID, placeID, start, finish)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Reservation) DeleteReservation(storage *postgres.Storage, reservationID int) error {
	const op = "storage.postgres.entities.reservation.DeleteReservation" // Имя текущей функции для логов и ошибок

	_, err := storage.DB.Exec(qrDeleteReservation, reservationID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Reservation) GetReservationsByUserID(storage *postgres.Storage, userID int) ([]Reservation, error) {
	const op = "storage.postgres.entities.reservation.GetReservationsByUserID" // Имя текущей функции для логов и ошибок

	qrResult, err := storage.DB.Query(qrGetReservationsByUserID, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var rs []Reservation
	for qrResult.Next() {
		var r Reservation
		if err := qrResult.Scan(&r.ReservationID, &r.PlaceID, &r.Start, &r.Finish); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		rs = append(rs, r)
	}

	return rs, nil
}
