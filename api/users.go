package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (s Server) usersProfile(c echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	user, err := s.db.User(c.Request().Context(), userID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return c.NoContent(http.StatusNotFound)
	case err != nil:
		return err
	}

	if len(user.Profile) == 0 {
		return c.NoContent(http.StatusNoContent)
	}

	return c.JSONBlob(http.StatusOK, user.Profile)
}
