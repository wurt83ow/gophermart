package controllers

type ExtController struct {
	log Log
}

func NewExtController(log Log) *ExtController {
	return &ExtController{
		log: log,
	}
}

func (c ExtController) GetData() (string, error) {
	return "data", nil
}
