package reservation

import (
	"fmt"
	"portal/internal/storage/postgres"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	// Получение актуальных мест 1. делаем все места "доступно" и вычитаем занятые 2. прибавляем занятые с пометкой "недостпуно"
	qrGetActualLockers = `(SELECT locker_id, "name", true AS is_available, 0 AS user_id, TIMESTAMP '0001-01-01 00:00:00' AS start, TIMESTAMP '0001-01-01 00:00:00' as finish FROM locker
						  EXCEPT
						  SELECT DISTINCT locker_id, "name", true AS is_available, 0, TIMESTAMP '0001-01-01 00:00:00', TIMESTAMP '0001-01-01 00:00:00' FROM locker_and_locker_reservation
						  WHERE ($1, $2) OVERLAPS ("start", finish))
						  UNION
						  (SELECT DISTINCT locker_id, "name", false AS is_available, user_id, start, finish FROM locker_and_locker_reservation
						  WHERE ($1, $2) OVERLAPS ("start", finish))
						  ORDER BY locker_id;`
	qrGetLockerReservationsByUserID       = `SELECT locker_reservation_id, locker_id, start, finish FROM locker_reservation WHERE user_id = $1 ORDER BY start DESC;`
	qrGetUserLockerReservationInDateRange = `SELECT locker_reservation_id FROM locker_reservation WHERE user_id = $1 AND (start, finish) OVERLAPS ($2, $3);`
	qrGetIsLockerAvailable                = `SELECT locker_reservation_id FROM locker_reservation WHERE locker_id = $1 AND (start, finish) OVERLAPS ($2, $3);`
	qrGetNameByLockerID                   = `SELECT name FROM locker WHERE locker_id = $1;`
	qrInsertLockerReservation             = `INSERT INTO locker_reservation (locker_id, start, finish, user_id) VALUES ($1, $3, $4, $2);`
	qrUpdateLockerReservation             = `UPDATE locker_reservation SET locker_id = $2, start = $3, finish = $4 WHERE locker_reservation_id = $1;`
	qrDeleteLockerReservation             = `DELETE FROM locker_reservation WHERE locker_reservation_id = $1;`
)

type Locker struct {
	LockerID int    `json:"locker_id,omitempty"`
	Name     string `json:"name,omitempty"`
}

func (l *Locker) GetLockerName(storage *postgres.Storage, lockerID int) error {
	const op = "storage.postgres.entities.reservation.GetLockerName"

	qrResult, err := storage.DB.Query(qrGetNameByLockerID, lockerID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer qrResult.Close()

	for qrResult.Next() {
		err = qrResult.Scan(&l.Name)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

type ActualLocker struct {
	Locker
	IsAvailable bool `json:"is_available"`
	UserID      int  `json:"user_id"`
	Start       int  `json:"start"`
	Finish      int  `json:"finish"`
}

func (al *ActualLocker) GetActualLockers(storage *postgres.Storage, start, finish time.Time) ([]ActualLocker, error) {
	const op = "storage.postgres.entities.reservation.GetActualLockers"

	qrResult, err := storage.DB.Query(qrGetActualLockers, start, finish)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer qrResult.Close()

	var als []ActualLocker

	for qrResult.Next() {
		var al ActualLocker
		if err := qrResult.Scan(&al.LockerID, &al.Name, &al.IsAvailable, &al.UserID, &al.Start, &al.Finish); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		als = append(als, al)
	}

	return als, nil
}

type LockerReservation struct {
	LockerReservationID int              `json:"locker_reservation_id,omitempty"`
	LockerID            int              `json:"locker_id,omitempty"`
	Start               pgtype.Timestamp `json:"start,omitempty"`
	Finish              pgtype.Timestamp `json:"finish,omitempty"`
	UserID              int              `json:"user_id,omitempty"`
}

func (r *LockerReservation) HasUserLockerReservationInDateRange(storage *postgres.Storage, userID int, start, finish string) (bool, error) {
	const op = "storage.postgres.entities.reservation.HasUserLockerReservationInDateRange" // Имя текущей функции для логов и ошибок

	qrResult, err := storage.DB.Query(qrGetUserLockerReservationInDateRange, userID, start, finish)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	defer qrResult.Close()

	// Проверка на пустой ответ
	if !qrResult.Next() {
		return false, nil
	}

	return true, nil
}

func (r *LockerReservation) InsertLockerReservation(storage *postgres.Storage, lockerID, userID int, start, finish string) error {
	const op = "storage.postgres.entities.reservation.InsertLockerReservation" // Имя текущей функции для логов и ошибок

	qrResult, err := storage.DB.Query(qrGetIsLockerAvailable, lockerID, start, finish)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer qrResult.Close()

	// Проверка на пустой ответ
	if qrResult.Next() {
		return fmt.Errorf("%s: locker is already taken", op)
	}

	_, err = storage.DB.Exec(qrInsertLockerReservation, lockerID, userID, start, finish)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (lr *LockerReservation) UpdateLockerReservation(storage *postgres.Storage, lockerReservationID, lockerID int, start, finish time.Time) error {
	const op = "storage.postgres.entities.reservation.UpdateLockerReservation" // Имя текущей функции для логов и ошибок

	_, err := storage.DB.Exec(qrUpdateLockerReservation, lockerReservationID, lockerID, start, finish)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (lr *LockerReservation) DeleteLockerReservation(storage *postgres.Storage, lockerReservationID int) error {
	const op = "storage.postgres.entities.reservation.DeleteLockerReservation" // Имя текущей функции для логов и ошибок

	_, err := storage.DB.Exec(qrDeleteLockerReservation, lockerReservationID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (lr *LockerReservation) GetLockerReservationsByUserID(storage *postgres.Storage, userID int) ([]LockerReservation, error) {
	const op = "storage.postgres.entities.reservation.GetLockerReservationsByUserID" // Имя текущей функции для логов и ошибок

	qrResult, err := storage.DB.Query(qrGetLockerReservationsByUserID, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer qrResult.Close()

	var lrs []LockerReservation
	for qrResult.Next() {
		var lr LockerReservation
		if err := qrResult.Scan(&lr.LockerReservationID, &lr.LockerID, &lr.Start, &lr.Finish); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		lrs = append(lrs, lr)
	}

	return lrs, nil
}
