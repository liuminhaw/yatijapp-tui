package authclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
)

type ErrInvalidToken struct {
	Msg string
}

func (e ErrInvalidToken) Error() string {
	return e.Msg
}

type Refresher func(ctx context.Context, refreshToken string) (Token, error)

func RefreshToken(serverURL string) Refresher {
	type refreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}
	type refreshResponse struct {
		AuthToken Token `json:"authentication_token"`
	}

	return func(ctx context.Context, refreshToken string) (Token, error) {
		req, err := NewJSONRequest(
			ctx,
			http.MethodPost,
			serverURL+"/v1/tokens/refresh",
			refreshRequest{
				RefreshToken: refreshToken,
			},
		)
		if err != nil {
			return Token{}, err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return Token{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			return Token{}, ErrInvalidToken{Msg: "failed to refresh token: " + resp.Status}
		}
		var token refreshResponse
		if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
			return Token{}, err
		}
		return Token{
			AccessToken:  token.AuthToken.AccessToken,
			RefreshToken: token.AuthToken.RefreshToken,
			SessionUUID:  token.AuthToken.SessionUUID,
		}, nil
	}
}

type AuthClient struct {
	Client    *http.Client
	Refresh   Refresher
	TokenPath string
	tokenMu   sync.RWMutex
}

func (c *AuthClient) SetToken(t Token) error {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	return TokenWrite(t, c.TokenPath)
}

func (c *AuthClient) GetToken() (Token, error) {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()

	token, err := TokenRead(c.TokenPath)
	if err != nil {
		return Token{}, ErrInvalidToken{Msg: "token is missing or invalid"}
	}
	return token, nil
}

func (c *AuthClient) ClearToken() error {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	return TokenDelete(c.TokenPath)
}

func (c *AuthClient) Do(req *http.Request) (*http.Response, error) {
	if c.Client == nil {
		c.Client = http.DefaultClient
	}

	retryRequest, err := cloneRequest(req)
	if err != nil {
		return nil, err
	}

	token, err := c.GetToken()
	if err != nil {
		return nil, err
	}
	setAuth(req, token.AccessToken)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusUnauthorized {
		return resp, nil
	}

	err = drainAndClose(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := c.refresh(req.Context()); err != nil {
		return nil, err
	}

	token, err = c.GetToken()
	if err != nil {
		return nil, err
	}
	setAuth(retryRequest, token.AccessToken)

	return c.Client.Do(retryRequest)
}

func (c *AuthClient) refresh(ctx context.Context) error {
	token, err := c.GetToken()
	if err != nil {
		return err
	}
	if token.RefreshToken == "" {
		return ErrInvalidToken{Msg: "no refresh token available"}
	}
	newToken, err := c.Refresh(ctx, token.RefreshToken)
	if err != nil {
		return err
	}
	if newToken.AccessToken == "" || newToken.RefreshToken == "" {
		return ErrInvalidToken{Msg: "invalid token received from refresher"}
	}

	if err := c.SetToken(newToken); err != nil {
		return err
	}
	return nil
}

func setAuth(req *http.Request, accessToken string) {
	if accessToken == "" {
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
}

func drainAndClose(rc io.ReadCloser) error {
	if rc == nil {
		return nil
	}
	_, _ = io.Copy(io.Discard, rc)
	return rc.Close()
}

func cloneRequest(req *http.Request) (*http.Request, error) {
	newReq := req.Clone(req.Context())
	if req.Body == nil {
		return newReq, nil
	}
	if req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		newReq.Body = body
		return newReq, nil
	}

	return nil, errors.New("cannot clone request with non-nil body and nil GetBody")
}

func NewJSONRequest(ctx context.Context, method, url string, payload any) (*http.Request, error) {
	var buf bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&buf).Encode(payload); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(buf.Bytes())), nil }

	return req, nil
}
