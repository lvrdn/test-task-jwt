package test

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"testing"

	jwt "github.com/golang-jwt/jwt/v5"
)

const key string = "!someKey321"

type AuthTestCase struct {
	GUID     string
	Path     string
	RT       *RefreshToken
	AT       *AccessToken
	Expected Response
}

type RefreshToken struct {
	CookieUpdate func(*http.Cookie) *http.Cookie
	NumCookieSt  int
}

type AccessToken struct {
	TokenUpdate func(string) string
	NumTokenSt  int
}

type Response struct {
	StatusCode             int
	GetCookie              bool
	GetHeaderAuthorization bool
}

func TestApp(t *testing.T) {

	cases := []*AuthTestCase{
		{ //0. получение access и refresh токенов, guid нет в запросе
			Path: "/api/auth",
			Expected: Response{
				StatusCode: http.StatusBadRequest,
			},
		},
		{ //1. получение access и refresh токенов, guid не валиден
			GUID: "qwerty",
			Path: "/api/auth?guid=",
			Expected: Response{
				StatusCode: http.StatusBadRequest,
			},
		},
		{ //2. получение access и refresh токенов, guid, которого нет в бд
			GUID: "5fccf5f2-63f1-4d61-a0a5-0920ff1d345c",
			Path: "/api/auth?guid=",
			Expected: Response{
				StatusCode: http.StatusBadRequest,
			},
		},
		{ //3. получение access и refresh токенов, guid, который есть в бд
			GUID: "2b421a0e-c5fa-47e3-a9d9-bc2fcf31ffe6",
			Path: "/api/auth?guid=",
			Expected: Response{
				StatusCode:             http.StatusOK,
				GetCookie:              true,
				GetHeaderAuthorization: true,
			},
		},
		{ //4. попытка обновить пару токенов, в запросе нет cookie с refresh token и header с access token
			Path: "/api/refresh",
			Expected: Response{
				StatusCode: http.StatusBadRequest,
			},
		},
		{ //5. попытка обновить пару токенов, нет cookie с refresh token
			Path: "/api/refresh",
			Expected: Response{
				StatusCode: http.StatusBadRequest,
			},
			AT: &AccessToken{
				NumTokenSt: 3,
			},
		},
		{ //6. попытка обновить пару токенов, в заголовке запроса нет access токена
			Path: "/api/refresh",
			Expected: Response{
				StatusCode: http.StatusBadRequest,
			},
			RT: &RefreshToken{
				NumCookieSt: 3,
			},
		},
		{ //7. попытка обновить пару токенов, неправильный refresh токен
			Path: "/api/refresh",
			Expected: Response{
				StatusCode: http.StatusUnauthorized,
			},
			AT: &AccessToken{
				NumTokenSt: 3, //в 3 тесте был получен access token
			},
			RT: &RefreshToken{
				NumCookieSt: 3, //в 3 тесте был получен refresh token
				CookieUpdate: func(cookie *http.Cookie) *http.Cookie {
					return &http.Cookie{
						Name:     cookie.Name,
						Value:    "test",
						HttpOnly: cookie.HttpOnly,
						Path:     cookie.Path,
						Expires:  cookie.Expires,
					}
				},
			},
		},
		{ //8. попытка обновить пару токенов, refresh токен не совпадает с хешем из бд
			Path: "/api/refresh",
			Expected: Response{
				StatusCode: http.StatusUnauthorized,
			},
			AT: &AccessToken{
				NumTokenSt: 3,
			},
			RT: &RefreshToken{
				NumCookieSt: 3,
				CookieUpdate: func(cookie *http.Cookie) *http.Cookie {

					refreshToken := cookie.Value
					tokenData := strings.Split(refreshToken, ".")
					tokenData[0] = "aaaaaaaaaaaaaaa"

					return &http.Cookie{
						Name:     cookie.Name,
						Value:    strings.Join(tokenData, "."),
						HttpOnly: cookie.HttpOnly,
						Path:     cookie.Path,
						Expires:  cookie.Expires,
					}
				},
			},
		},
		{ //9. попытка обновить пару токенов, неправильный matching key у access tokena, который не совпадет с refresh токеном
			Path: "/api/refresh",
			Expected: Response{
				StatusCode: http.StatusUnauthorized,
			},
			AT: &AccessToken{
				NumTokenSt: 3,
				TokenUpdate: func(token string) string {
					claims := jwt.MapClaims{}
					_, _ = jwt.ParseWithClaims(
						token,
						&claims,
						func(token *jwt.Token) (interface{}, error) {
							return []byte(key), nil
						},
					)

					claims["matching_key"] = "new_key" //новый matching key
					newToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
					signedRefreshToken, _ := newToken.SignedString([]byte(key))
					return signedRefreshToken
				},
			},
			RT: &RefreshToken{
				NumCookieSt: 3,
			},
		},
		{ //10. попытка обновить пару токенов, у access токена изменен payload, т.е. будет ошибка при проверке подписи
			Path: "/api/refresh",
			Expected: Response{
				StatusCode: http.StatusUnauthorized,
			},
			AT: &AccessToken{
				NumTokenSt: 3,
				TokenUpdate: func(token string) string {
					s := strings.Split(token, ".")
					payload, _ := base64.StdEncoding.DecodeString(s[1])
					newPayload := strings.ReplaceAll(string(payload), "authApp", "unknownApp") //изменение payload
					s[1] = base64.StdEncoding.EncodeToString([]byte(newPayload))

					return strings.Join(s, ".")
				},
			},
			RT: &RefreshToken{
				NumCookieSt: 3,
			},
		},
		{ //11. обновление пары токенов, правильный refresh и access токены
			Path: "/api/refresh",
			Expected: Response{
				StatusCode:             http.StatusOK,
				GetCookie:              true,
				GetHeaderAuthorization: true,
			},
			AT: &AccessToken{
				NumTokenSt: 3,
			},
			RT: &RefreshToken{
				NumCookieSt: 3,
			},
		},
		{ //12. попытка обновить пару токенов, используя старый refresh токен (полученный в кейсе 1)
			Path: "/api/refresh",
			Expected: Response{
				StatusCode: http.StatusUnauthorized,
			},
			AT: &AccessToken{
				NumTokenSt: 11,
			},
			RT: &RefreshToken{
				NumCookieSt: 3,
			},
		},
		{ //13. имитация обновления пары токенов с другого ip адреса
			Path: "/api/refresh",
			Expected: Response{
				StatusCode:             http.StatusOK,
				GetCookie:              true,
				GetHeaderAuthorization: true,
			},
			AT: &AccessToken{
				NumTokenSt: 11,
				TokenUpdate: func(token string) string {
					claims := jwt.MapClaims{}
					_, _ = jwt.ParseWithClaims(
						token,
						&claims,
						func(token *jwt.Token) (interface{}, error) {
							return []byte(key), nil
						},
					)

					claims["ip"] = "255.255.255.255" //новый ip
					newToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
					signedRefreshToken, _ := newToken.SignedString([]byte(key))

					return signedRefreshToken
				},
			},
			RT: &RefreshToken{
				NumCookieSt: 11,
			},
		},
	}

	cookiesStorage := make(map[int]*http.Cookie) //ключ - номер тест-кейса, в котором была получена cookie с refresh token
	accessTokenStorage := make(map[int]string)   //ключ - номер тест-кейса, в котором был получен access token

	ctx, cancel := context.WithCancel(context.Background())
	app := New(ctx) //запуск тестового сервера
	defer app.server.Close()

	for i, testCase := range cases {
		client := &http.Client{}
		req, err := http.NewRequest(http.MethodGet, app.server.URL+testCase.Path+testCase.GUID, nil)
		if err != nil {
			t.Fatalf("make request error, num case: [%d], error msg: [%s]", i, err.Error())
		}

		if testCase.AT != nil {
			accessToken := accessTokenStorage[testCase.AT.NumTokenSt]
			if testCase.AT.TokenUpdate != nil {
				accessToken = testCase.AT.TokenUpdate(accessToken)
			}
			req.Header.Set("Authorization", accessToken)
		}

		if testCase.RT != nil {
			cookieWithRefreshToken := cookiesStorage[testCase.RT.NumCookieSt]
			if testCase.RT.CookieUpdate != nil {
				cookieWithRefreshToken = testCase.RT.CookieUpdate(cookieWithRefreshToken)
			}
			req.AddCookie(cookieWithRefreshToken)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("get response error, num case: [%d], error msg: [%s]", i, err.Error())
		}
		if resp.StatusCode != testCase.Expected.StatusCode {
			t.Fatalf("unexpected status code: num case: [%d], expected [%d], got [%d]", i, testCase.Expected.StatusCode, resp.StatusCode)
		}

		if testCase.Expected.GetCookie {
			if len(resp.Cookies()) != 1 {
				t.Fatalf("response must have one cookie with refresh token in this num case [%d]", i)
			} else {
				cookiesStorage[i] = resp.Cookies()[0]
			}
		}

		if testCase.Expected.GetHeaderAuthorization {
			accessToken := resp.Header.Get("Authorization")
			if accessToken == "" {
				t.Fatalf("response must have header with access token in this num case [%d], but header Authorization is empty", i)
			} else {

				accessTokenStorage[i] = accessToken

			}
		}
		resp.Body.Close()
	}

	cancel()
	app.WaitClosingProcceses()

}
