package friendlycaptcha

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type VerifyRequest struct {
	Solution string `json:"solution"`
	Secret   string `json:"secret,omitempty"`
	Sitekey  string `json:"sitekey,omitempty"`
}

type VerifyResponse struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors,omitempty"`
	Details *string  `json:"details,omitempty"`
}

// A client for the Friendly Captcha API, see also the API docs at https://docs.friendlycaptcha.com
type Client struct {
	APIKey        string
	Sitekey       string
	SiteverifyURL string
	// If Strict is set to true only strictly verified captcha solutions will be allowed.
	// By default it is false: the accept value will be true when for instance the Friendly Captcha API
	// could not be reached.
	Strict bool
}

const SolutionFormFieldName = "frc-captcha-solution"
const defaultSiteVerifyAPIURL = "https://api.friendlycaptcha.com/api/v1/siteverify"

// Could not create the request body (i.e. JSON marshal it), this should never happen but if it does then probably
// the captcha solution value is really weird - let's not accept the verification
var ErrCreatingVerificationRequest = errors.New("could not create verification request body")

// The POST request to the Friendly Captcha API could not be completed for some reason.
var ErrVerificationRequest = errors.New("verification request failed talking to Friendly Captcha API")

// This error signifies a non-200 response from the server. Usually this means that your API key was wrong.
// You should notify yourself if this happens, but it's usually still a good idea to accept the captcha even though
// we were unable to verify it: we don't want to lock users out.
var ErrVerificationFailedDueToClientError = errors.New("verification request failed due to a client error (check your credentials)")

func NewClient(apiKey string, sitekey string) Client {
	return Client{
		APIKey:        apiKey,
		Sitekey:       sitekey,
		SiteverifyURL: defaultSiteVerifyAPIURL,
	}
}

func (frc Client) CheckCaptchaSolution(ctx context.Context, captchaSolution string) (shouldAccept bool, err error) {
	reqBody := VerifyRequest{
		Solution: captchaSolution,
		Secret:   frc.APIKey,
		Sitekey:  frc.Sitekey,
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrCreatingVerificationRequest, err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", frc.SiteverifyURL, bytes.NewReader(reqBodyJSON))
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrCreatingVerificationRequest, err)
	}

	req.Header.Set("X-Frc-Sdk", "go-sdk-0.1.1")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return !frc.Strict, fmt.Errorf("%w: %v", ErrVerificationRequest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		// Intentionally let this through, it's probably a problem in our credentials
		return !frc.Strict, fmt.Errorf("%w [status %d]: %s", ErrVerificationFailedDueToClientError, resp.StatusCode, b)
	}

	decoder := json.NewDecoder(resp.Body)
	var vr VerifyResponse
	err = decoder.Decode(&vr)
	if err != nil {
		// Despite the error we will let this through - it must be a problem with the Friendly Captcha API.
		return !frc.Strict, fmt.Errorf("%w: %v", ErrVerificationRequest, err)
	}

	if !vr.Success {
		return false, nil
	}

	return true, nil
}
