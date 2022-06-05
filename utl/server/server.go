package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/swagftw/covax19-blockchain/utl/server/fault"
)

// StartHTTPServer starts the HTTP server.
func StartHTTPServer(e *echo.Echo) error {
	return e.Start(":9090")
}

func InitEcho() *echo.Echo {
	// InitEcho initializes the echo instance.
	e := echo.New()
	// JwtMiddleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.HTTPErrorHandler = fault.ErrorHandler

	return e
}

// SendRequest sends a request to the given URL.
func SendRequest(method string, url string, payload interface{}) (interface{}, error) {
	var client http.Client

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		log.Println(err)

		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		log.Println(err)

		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(response.Body)

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)

		return nil, err
	}

	log.Println(string(body))

	faultErr := make(map[string]*fault.HTTPError)

	if response.StatusCode >= http.StatusBadRequest {
		_ = json.Unmarshal(body, &faultErr)

		return nil, faultErr["error"]
	}

	var respPayload map[string]interface{}

	err = json.Unmarshal(body, &respPayload)
	if err != nil {
		log.Println(err)

		return nil, err
	}

	return respPayload, nil
}

func ToGoContext(c echo.Context) context.Context {
	type key string

	var newKey key = "key"

	return context.WithValue(c.Request().Context(), newKey, "value")
}
