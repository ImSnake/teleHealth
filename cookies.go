package main

import "net/http"

func clearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   sessionName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}

func setCookie(w http.ResponseWriter, cookieName, cookieValue string) {

	formCookie := &http.Cookie{
		Name:     cookieName,
		Value:    cookieValue,
		Path:     "/",
		MaxAge:   300,                     // 5 минут

		// TODO заменить на true
		Secure:   false,                   // yet 'false' as TLS is not used

		// TODO заменить на true
		HttpOnly: false,                    // 'true' secures from XSS attacks

		// TODO заменить на http.SameSiteStrictMode
		SameSite: http.SameSiteStrictMode, // base CSRF attack protection
	}

	http.SetCookie(w, formCookie)
}
