package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/wurt83ow/gophermart/internal/models"
	"go.uber.org/zap"
)

type ExtController struct {
	log Log
}

func NewExtController(log Log) *ExtController {
	return &ExtController{
		log: log,
	}
}

func (c ExtController) GetData() (*models.ExtRespOrder, error) {

	url := "http://localhost:8082/api/orders/356477"
	resp, err := http.Get(url)
	if err != nil {
		// we will get an error at this stage if the request fails, such as if the
		// requested URL is not found, or if the server is not reachable.
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// if we want to check for a specific status code, we can do so here
	// for example, a successful request should return a 200 OK status
	if resp.StatusCode != http.StatusOK {
		// if the status code is not 200, we should log the status code and the
		// status string, then exit with a fatal error
		code := zap.String("code", strconv.Itoa(resp.StatusCode))
		status := zap.String("code", resp.Status)
		c.log.Info("status code error: ", code, status)
	}

	respOrd := models.ExtRespOrder{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&respOrd); err != nil {
		c.log.Info("cannot decode request JSON body: ", zap.Error(err))

		return nil, err
	}

	fmt.Println("7777777777777777777777777", respOrd.Status)
	return &respOrd, nil
}
