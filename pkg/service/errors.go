package service

import (
	"encoding/json"
	"net/http"
)

type SpeakerbobError string

func (e SpeakerbobError) Error() string {
	return string(e)
}

type NotAcceptableError struct {
	SpeakerbobError
}

func NewNotAcceptableError(msg string) NotAcceptableError {
	return NotAcceptableError{
		SpeakerbobError(msg),
	}
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func WriteErrorResponse(w http.ResponseWriter, err error) {
	resp := errorResponse{
		Code:    http.StatusInternalServerError,
		Message: "An unexpected error has occurred.",
	}

	if err == nil {
		return
	}

	switch err.(type) {
	case SpeakerbobError:
		resp.Message = err.Error()
		break
	case NotAcceptableError:
		resp.Code = http.StatusNotAcceptable
		resp.Message = err.Error()
		break
	}

	w.WriteHeader(resp.Code)
	_ = json.NewEncoder(w).Encode(resp)
}
