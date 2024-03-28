package handler

import (
	"github.com/isaquerr25/go-templ-htmx/model"
	"github.com/isaquerr25/go-templ-htmx/view/user"
	"github.com/labstack/echo/v4"
)

type UserHandler struct{}

func (h *UserHandler) HandleUserShow(c echo.Context) error {
	u := model.User{
		Email: "a@gmail.com",
	}
	return render(c, user.Show(u))

}
