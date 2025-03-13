package phoneBook

import (
	"encoding/json"
	"log/slog"
	"net/http"
	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"
	"portal/internal/storage/postgres"
	"portal/internal/storage/postgres/entities/user"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type EmployeeInfo struct {
	FullName   string `json:"full_name"`
	Position   string `json:"position"`
	Department string `json:"department"`
	Mail       string `json:"mail"`
	Mobile     string `json:"mobile"`
}

type Response struct {
	resp.Response
	EmployeesInfo []EmployeeInfo `json:"employees"`
}

func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.phoneBook.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var u user.User
		var eis []EmployeeInfo
		us, err := u.GetAllUsersInfo(storage)
		if err != nil {
			log.Error("failed to get users info", sl.Err(err))
			w.WriteHeader(422)
			render.JSON(w, r, resp.Error("failed to get users info"))
			return
		}

		for _, u := range us {
			eis = append(eis, EmployeeInfo{FullName: u.FullName, Position: u.Position, Department: u.Department, Mail: u.Mail, Mobile: u.Mobile})
		}

		log.Info("users gotten")

		responseOK(w, r, log, eis)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, log *slog.Logger, eis []EmployeeInfo) {
	response, err := json.Marshal(Response{
		Response:      resp.OK(),
		EmployeesInfo: eis,
	})
	if err != nil {
		log.Error("failed to process response", sl.Err(err))
		w.WriteHeader(500)
		render.JSON(w, r, resp.Error("failed to process response"))
		return
	}

	render.Data(w, r, response)
}
