/*
Copyright 2002 Adevinta
*/

package api

import (
	"net/http"

	"github.com/adevinta/vulcan-jira-api/pkg/issues"
	"github.com/adevinta/vulcan-jira-api/pkg/model"
	"github.com/labstack/echo/v4"
)

func response(c echo.Context, httpStatus int, data interface{}, dataType string, p ...issues.Pagination) error {
	if data == nil {
		return c.NoContent(http.StatusNoContent)
	}

	resp := map[string]interface{}{}

	// We check if the variadic argument is present.
	if len(p) > 0 {
		// We only use the first element, as we expect only one.
		more := p[0].Total > p[0].Offset+p[0].Limit

		pagination := Pagination{
			Limit:  p[0].Limit,
			Offset: p[0].Offset,
			Total:  p[0].Total,
			More:   more,
		}

		if p[0].Offset > p[0].Total {
			return echo.NewHTTPError(http.StatusNotFound, ErrPageNotFound.Error())
		}

		resp["pagination"] = pagination
	}

	resp[dataType] = data

	return c.JSON(httpStatus, resp)
}

// GetIssue returns a JSON containing a specific issue.
func (api *API) GetIssue(c echo.Context) error {
	id := c.Param("id")

	// We need here
	issue, err := api.issueTracking.GetIssue(id)
	if err != nil {
		return err
	}

	if issue.ID == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	return response(c, http.StatusOK, issue, "issue")
}

// CreateIssue creates an issue and returns a JSON containing the new issue.
func (api *API) CreateIssue(c echo.Context) error {
	issue := new(model.Issue)
	if err := c.Bind(issue); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	issue, err := api.issueTracking.CreateIssue(issue)
	if err != nil {
		return err
	}
	if issue.ID == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	return response(c, http.StatusOK, issue, "issue")
}

// FixIssue updates an issue until a "done" state and returns a JSON containing the new issue.
func (api *API) FixIssue(c echo.Context) error {
	id := c.Param("id")

	transitions, err := api.issueTracking.GetTransitions(id)
	if err != nil {
		return err
	}
	for _, transition := range *transitions {
		if transition.ToName == "Fixed" {
			return nil
		}
	}

	issue, err := api.issueTracking.FixIssue(id)
	if err != nil {
		return err
	}
	if issue.ID == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	return response(c, http.StatusOK, issue, "issue")
}
