package handler

import (
	"app/internal/sender"
	"app/internal/storage"
	"app/internal/tokens"
	"app/pkg/logger"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type resp map[string]interface{}

type handler struct {
	storage      storage.Storage
	emailSender  sender.EmailSender
	tokenManager tokens.TokenManager
}

func NewHandler(
	st storage.Storage,
	emailSender sender.EmailSender,
	tokenManager tokens.TokenManager,
) *handler {
	return &handler{
		storage:      st,
		emailSender:  emailSender,
		tokenManager: tokenManager,
	}
}

func (h *handler) issue(w http.ResponseWriter, r *http.Request) {

	const methodPtr string = "handler.issue"

	guid := r.FormValue("guid")
	_, err := uuid.Parse(guid)

	if guid == "" || err != nil {
		errMsg := "request must have query param valid guid"
		response := resp{
			"error":  errMsg,
			"path":   r.URL.Path,
			"method": r.Method,
			"time":   time.Now().String(),
		}
		responseData, err := json.Marshal(response)
		if err != nil {
			logger.Error("marshal response error", methodPtr, "error", err.Error(), "response msg error", errMsg)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(responseData)
		logger.Info("response with error sended", methodPtr, "response", string(responseData), "status code", http.StatusBadRequest)
		return
	}

	id, err := h.storage.CheckGUID(guid)
	if err != nil {
		if err == sql.ErrNoRows {
			errMsg := "unknown guid"
			response := resp{
				"error":  errMsg,
				"path":   r.URL.Path,
				"method": r.Method,
				"time":   time.Now().String(),
			}
			responseData, err := json.Marshal(response)
			if err != nil {
				logger.Error("marshal response error", methodPtr, "error", err.Error(), "response msg error", errMsg)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(responseData)
			logger.Info("response with error sended", methodPtr, "response", string(responseData), "status code", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error("check guid in db failed", methodPtr, "error", err.Error())
		return
	}

	refreshToken, err := h.tokenManager.CreateRefreshToken(id)
	if err != nil {
		logger.Error("get refresh token failed", methodPtr, "error", err.Error(), "refresh token", refreshToken.Value)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.storage.AddNewRefreshToken(id, refreshToken.HashedValue, refreshToken.ExpDate)
	if err != nil {
		logger.Error("add new refresh token to storage failed", methodPtr,
			"error", err.Error(),
			"user id", id,
			"refresh token", refreshToken.Value,
			"hashed refresh token", refreshToken.HashedValue,
			"exp date", refreshToken.ExpDate,
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	signedAccessToken, err := h.tokenManager.CreateSignedAccessToken(id, r.RemoteAddr, refreshToken.MatchingKey)
	if err != nil {
		logger.Error("create signed access token error", methodPtr,
			"error", err.Error(),
			"user id", id,
			"remote addr", r.RemoteAddr,
			"matchingKey", refreshToken.MatchingKey,
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken.Value,
		HttpOnly: true,
		Path:     "/api/refresh",
		Expires:  refreshToken.ExpDate,
	})

	w.Header().Set("Authorization", signedAccessToken)
	logger.Info("access, refresh tokens successfully created", methodPtr, "access token", signedAccessToken, "refresh token", refreshToken.Value, "user id", id)
}

func (h *handler) refresh(w http.ResponseWriter, r *http.Request) {

	const methodPtr string = "handler.refresh"

	cookie, err := r.Cookie("refresh_token")
	accessTokenFromReq := r.Header.Get("Authorization")

	if err != nil || accessTokenFromReq == "" {
		errMsg := "request must have cookie with refresh token and access token in header Authorization"
		response := resp{
			"error":  errMsg,
			"path":   r.URL.Path,
			"method": r.Method,
			"time":   time.Now().String(),
		}
		responseData, err := json.Marshal(response)
		if err != nil {
			logger.Error("marshal response error", methodPtr, "error", err.Error(), "response msg error", errMsg)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(responseData)
		logger.Info("response with error sended", methodPtr, "response", string(responseData), "status code", http.StatusBadRequest)
		return
	}

	refreshTokenFromReq := cookie.Value

	refreshToken, err := h.tokenManager.ParseRefreshTokenFromReq(refreshTokenFromReq)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		logger.Info("bad refresh token", methodPtr, "error", err.Error(), "refresh token from req", refreshTokenFromReq, "status code", http.StatusUnauthorized)
		return
	}

	accessToken, err := h.tokenManager.ParseAccessTokenFromReq(accessTokenFromReq)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		logger.Info("bad access token", methodPtr, "error", err.Error(), "access token from req", accessTokenFromReq, "status code", http.StatusUnauthorized)
		return
	}

	//проверка на соответствие refresh и access токенов по matching key
	if accessToken.MatchingKey != refreshToken.MatchingKey {
		w.WriteHeader(http.StatusUnauthorized)
		logger.Info("matching keys mismatch", methodPtr, "access token matching key", accessToken.MatchingKey, "refresh token matching key", refreshToken.MatchingKey, "status code", http.StatusUnauthorized)
		return
	}

	hashedRefreshToken, expDate, guid, err := h.storage.GetRefreshTokenData(accessToken.UserID)
	if err != nil {
		//TODO если не нашел ничего в базе
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error("get hashed refresh token and exp date failed", methodPtr, "error", err.Error(), "user_id", accessToken.UserID)
		return
	}

	//проверка срока жизни refresh токена
	if time.Now().After(expDate) {
		w.WriteHeader(http.StatusUnauthorized)
		logger.Info("refresh token expired", methodPtr, "user_id", accessToken.UserID, "status code", http.StatusUnauthorized)
		return
	}

	//проверка на соответствие refresh токена из запроса и из базы
	if !h.tokenManager.CompareHashAndToken(hashedRefreshToken, refreshToken.Value) {
		w.WriteHeader(http.StatusUnauthorized)
		logger.Info("refresh token and hash mismatch", methodPtr, "user_id", accessToken.UserID, "status code", http.StatusUnauthorized)
		return
	}

	//отправление warning письма юзеру о доступе к его данным с другого устройства
	if r.RemoteAddr != accessToken.IP {
		warnMsg := fmt.Sprintf("unknown ip get access to refresh operation: unknown ip: [%s], expected ip: [%s], guid: [%s]\n", r.RemoteAddr, accessToken.IP, guid)
		logger.Warn(warnMsg, methodPtr)

		msg := fmt.Sprintf("WARNING, somebody get access to your data\nip: [%s]\nuser-agent: [%s]\nIf this is you, ignore this message\n", r.RemoteAddr, r.UserAgent())
		err := h.emailSender.Send(guid, msg)
		if err != nil {
			log.Printf("send warning message error: [%s], refresh operation stopped\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("send warning message failed", methodPtr, "error", err.Error(), "guid", guid, "msg", msg)
			return
		} else {
			infoMsg := fmt.Sprintf("warning message succesfully sended to [%s]\n", guid)
			logger.Info(infoMsg, methodPtr)
		}
	}

	newRefreshToken, err := h.tokenManager.CreateRefreshToken(accessToken.UserID)
	if err != nil {
		logger.Error("get new refresh token failed", methodPtr, "error", err.Error(), "refresh token", newRefreshToken.Value)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.storage.AddNewRefreshToken(accessToken.UserID, newRefreshToken.HashedValue, newRefreshToken.ExpDate)
	if err != nil {
		log.Printf("error with add new hashed refresh token to db: [%s]\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//создание нового подписанного access токена
	newSignedAccessToken, err := h.tokenManager.CreateSignedAccessToken(accessToken.UserID, r.RemoteAddr, newRefreshToken.MatchingKey)
	if err != nil {
		logger.Error("create signed access token error", methodPtr,
			"error", err.Error(),
			"user id", accessToken.UserID,
			"remote addr", r.RemoteAddr,
			"matchingKey", newRefreshToken.MatchingKey,
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken.Value,
		HttpOnly: true,
		Path:     "/api/refresh",
		Expires:  newRefreshToken.ExpDate,
	})

	w.Header().Set("Authorization", newSignedAccessToken)
	logger.Info("access, refresh tokens successfully created", methodPtr, "access token", newSignedAccessToken, "refresh token", refreshToken.Value, "user id", accessToken.UserID)
}
