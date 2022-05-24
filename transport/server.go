package transport

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	ErrBadRequest = errors.New("bad request")
)

// StartHTTPServer starts the HTTP server.
func StartHTTPServer(e *echo.Echo) error {
	return errors.Wrap(e.Start(":9090"), "failed to start HTTP server")
}

// InitEcho initializes the echo instance.
func InitEcho() *echo.Echo {
	e := echo.New()
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	return e
}

func sendRequest(method string, url string, payload interface{}) (interface{}, error) {
	var client http.Client

	body, err := json.Marshal(payload)
	if err != nil {
		err = errors.Wrap(err, "failed to marshal payload")
		log.Println(err)

		return nil, err
	}

	request, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		log.Println(errors.Wrap(err, "failed to create request"))

		return nil, errors.Wrap(err, "failed to create request")
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		err = errors.Wrap(err, "failed to send request")
		log.Println(err)

		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Println(errors.Wrap(err, "failed to close response body"))
		}
	}(response.Body)

	if response.StatusCode >= http.StatusBadRequest {
		return nil, ErrBadRequest
	}

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		err = errors.Wrap(err, "failed to read response body")
		log.Println(err)

		return nil, err
	}

	log.Println(string(body))

	var respPayload map[string]interface{}

	err = json.Unmarshal(body, &respPayload)
	if err != nil {
		err = errors.Wrap(err, "failed to unmarshal response body")
		log.Println(err)

		return nil, err
	}

	return respPayload, nil
}
