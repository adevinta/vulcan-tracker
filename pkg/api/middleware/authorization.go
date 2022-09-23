/*
Copyright 2022 Adevinta
*/

package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/adevinta/vulcan-jira-api/pkg/issues"
	echo "github.com/labstack/echo/v4"
)

const (
	issuesOpPrefix = "/issues/"

	tagAuthSchemaPrefix = "TAG tag="
)

var (
	errUnauthorized = errors.New("unauthorized")
	errForbidden    = errors.New("forbidden")
	errNotFound     = errors.New("resource not found")
)

// Authorization returns a new authorization middleware func
// by using the input authorizer.
func Authorization(auth Authorizer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			err = auth.Authorize(c.Request())
			if err != nil {
				c.Response().WriteHeader(errToStatus(err))
				return
			}
			return next(c)
		}
	}
}

// Authorizer represents an authorization
// mechanism applied to an HTTP request.
type Authorizer interface {
	Authorize(r *http.Request) error
}

// TagAuthorizer performs authorization based
// on the TAG scheme for the authorization header.
type TagAuthorizer struct {
	issueTracking issues.IssueTracking
	log           echo.Logger
}

// NewTagAuthorizer builds a new tag based authorizer.
// This type of authorizer matches the tag passed in through
// the authorization header against the tag associated with the
// resource that the request tries to modify.
func NewTagAuthorizer(issueTracking issues.IssueTracking, log echo.Logger) *TagAuthorizer {
	return &TagAuthorizer{
		issueTracking,
		log,
	}
}

// Authorize authorizes HTTP request by verifying resource ownership.
// It does so by retrieving the tag included in the http request's
// authorization header and comparing it with the tags associated
// with the resource that the request is trying to modify.
// If request tag is among the ones associated with the resource,
// then action is granted. Otherwise it is denied.
func (a *TagAuthorizer) Authorize(r *http.Request) error {
	if isReadMethod(r.Method) {
		// Read HTTP methods do not
		// require authorization
		return nil
	}

	entityID := parseEntityID(r.URL.Path)
	if r.Method == http.MethodPost && entityID == "" {
		// Authorize creation operations.
		// E.g.: POST /findings
		return nil
	}
	return nil

}

func parseEntityID(path string) string {
	if pathParts := strings.Split(path, "/"); len(pathParts) > 2 {
		return pathParts[2]
	}
	return ""
}

func errToStatus(err error) int {
	var status int

	if err == nil {
		status = http.StatusOK
	} else if errors.Is(err, errUnauthorized) {
		status = http.StatusUnauthorized
	} else if errors.Is(err, errForbidden) {
		status = http.StatusForbidden
	} else if errors.Is(err, errNotFound) {
		status = http.StatusNotFound
	} else {
		status = http.StatusInternalServerError
	}

	return status
}

func isReadMethod(method string) bool {
	return method == http.MethodGet ||
		method == http.MethodHead ||
		method == http.MethodOptions
}

func contains(slice []string, s string) bool {
	for _, ss := range slice {
		if ss == s {
			return true
		}
	}
	return false
}
