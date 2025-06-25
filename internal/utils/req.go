package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/nihal-ramaswamy/RunnerIO/internal/dto"
	"go.uber.org/zap"
)

func GetUserInfo(issuerURL *url.URL, token string, log *zap.Logger) (*dto.UserData, error) {
	userInfoUrl, err := url.JoinPath(issuerURL.String(), "/userinfo")
	if err != nil {
		log.Fatal("Failed to join the user info url", zap.Error(err))
		return nil, err
	}

	client := &http.Client{}
	method := http.MethodGet

	req, err := http.NewRequest(method, userInfoUrl, nil)

	if err != nil {
		log.Fatal("Failed to create the request", zap.Error(err))
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("authorization", "Bearer "+token)
	res, err := client.Do(req)

	if err != nil {
		log.Fatal("Failed to make the request", zap.Error(err))
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal("Failed to read the response body", zap.Error(err))
		return nil, err
	}

	var userData dto.UserData
	if err := json.Unmarshal([]byte(body), &userData); err != nil {
		log.Error("Failed to unmarshal the user data", zap.Error(err))
		return nil, err
	}

	return &userData, nil
}

func GetMgmtPostResponse(mgmtClientId, mgmtClientSecret, mgmtAudienceApi, mgmtTokenUrl string, log *zap.Logger) (*dto.MgmtPostResponse, error) {
	client := &http.Client{}
	method := http.MethodPost
	mgmtBody := dto.NewMgmtPostRequest(mgmtClientId, mgmtClientSecret, mgmtAudienceApi)
	bbody, err := json.Marshal(mgmtBody)
	if err != nil {
		log.Fatal("Failed to marshal the body", zap.Error(err))
		return nil, err
	}

	req, err := http.NewRequest(method, mgmtTokenUrl, bytes.NewBuffer(bbody))
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)

	if err != nil {
		log.Fatal("Failed to make the request", zap.Error(err))
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal("Failed to read the response body", zap.Error(err))
		return nil, err
	}

	var mgmtPostResponse dto.MgmtPostResponse
	if err := json.Unmarshal([]byte(string(body)), &mgmtPostResponse); err != nil {
		log.Error("Failed to unmarshal the user data", zap.Error(err))
		return nil, err
	}

	return &mgmtPostResponse, nil
}

func GetUserPermissions(mgmtPermissionUrl, mgmtTokenUrl, mgmtClientId, mgmtClientSecret, mgmtAudienceApi, access_token string, log *zap.Logger) ([]dto.UserPermission, error) {
	client := &http.Client{}
	method := http.MethodGet

	req, err := http.NewRequest(method, mgmtPermissionUrl, nil)

	if err != nil {
		log.Fatal("Failed to create the request", zap.Error(err))
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("authorization", "Bearer "+access_token)
	res, err := client.Do(req)

	if err != nil {
		log.Fatal("Failed to make the request", zap.Error(err))
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal("Failed to read the response body", zap.Error(err))
		return nil, err
	}

	var userPermissions []dto.UserPermission
	if err := json.Unmarshal([]byte(string(body)), &userPermissions); err != nil {
		log.Error("Failed to unmarshal the user data", zap.Error(err))
		return nil, err
	}

	return userPermissions, nil
}
