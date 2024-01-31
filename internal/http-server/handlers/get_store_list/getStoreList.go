package getStoreList

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/gofiber/fiber/v2/log"

	resp "portal/internal/lib/api/response"
)

type Response struct {
	resp.Response
	StoreList string `json:"store_list"`
}

type StoreListGetter interface {
	GetStoreList() (*string, error)
}

func New(storeListGetter StoreListGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.getStoreList.New"

		storeList, err := storeListGetter.GetStoreList()
		if err != nil {
			log.Error("failed to get store list")

			render.JSON(w, r, resp.Error("failed to get store list"))

			return
		}

		log.Info("store list gotten")

		responseOK(w, r, *storeList)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, storeList string) {
	render.JSON(w, r, Response{
		Response:  resp.OK(),
		StoreList: storeList,
	})
}
