package publish

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Timspects/roblox-kartfr-asset-reuploader-modified-by-timspects/internal/roblox"
)

var UploadImageErrors = struct {
	ErrModerated        error
	ErrTokenInvalid     error
	ErrNotAuthenticated error
}{
	ErrModerated:        errors.New("moderated name or description"),
	ErrTokenInvalid:     errors.New("XSRF token validation failed"),
	ErrNotAuthenticated: errors.New("user is not authenticated"),
}

type uploadImageRequest struct {
	Name              string  `json:"name"`
	File              string  `json:"file"`
	GroupID           int64   `json:"groupId,omitempty"`
	PaymentSource     string  `json:"paymentSource,omitempty"`
	EstimatedFileSize int64   `json:"estimatedFileSize"`
	EstimatedDuration float64 `json:"estimatedDuration"`
	AssetPrivacy      int32   `json:"assetPrivacy"`
}

type publishImageResponse struct {
	ID     int64  `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Errors []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

func newUploadImageRequest(name string, data *bytes.Buffer, groupID ...int64) (*http.Request, error) {
	var buffer bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buffer)
	size := int64(data.Len())
	if _, err := io.Copy(encoder, data); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}

	body := uploadImageRequest{
		Name:              name,
		File:              buffer.String(),
		EstimatedFileSize: size,
	}
	if len(groupID) > 0 {
		body.GroupID = groupID[0]
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://publish.roblox.com/v1/assets/upload", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "RobloxStudio/WinInet")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func NewUploadImageHandler(c *roblox.Client, name string, data *bytes.Buffer, groupID ...int64) (func() (*publishImageResponse, error), error) {
	req, err := newUploadImageRequest(name, data, groupID...)
	if err != nil {
		return func() (*publishImageResponse, error) { return nil, nil }, err
	}

	return func() (*publishImageResponse, error) {
		req.AddCookie(&http.Cookie{
			Name:  ".ROBLOSECURITY",
			Value: c.Cookie,
		})
		req.Header.Set("x-csrf-token", c.GetToken())

		resp, err := c.DoRequest(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var response publishImageResponse
		json.NewDecoder(resp.Body).Decode(&response)

		switch resp.StatusCode {
		case http.StatusOK:
			return &response, nil
		case http.StatusBadRequest:
			if response.Errors == nil {
				return nil, errors.New(resp.Status)
			}

			message := response.Errors[0].Message
			if message == "Image name or description is moderated." {
				req, _ = newUploadImageRequest("[Censored]", data, groupID...)
				return nil, UploadImageErrors.ErrModerated
			}

			return nil, errors.New(message)
		case http.StatusUnauthorized:
			if response.Errors == nil {
				return nil, errors.New(resp.Status)
			}

			message := response.Errors[0].Message
			if message == "User is not authenticated" {
				return nil, UploadImageErrors.ErrNotAuthenticated
			}

			return nil, errors.New(message)
		case http.StatusForbidden:
			c.SetToken(resp.Header.Get("x-csrf-token"))
			return nil, UploadImageErrors.ErrTokenInvalid
		default:
			if response.Errors == nil {
				return nil, errors.New(resp.Status)
			}

			return nil, errors.New(response.Errors[0].Message)
		}
	}, nil
}
