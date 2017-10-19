package handlers

import (
	"net/http"
)

//CheckAuth check the authorization
func CheckAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(HXAPPID) == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)

		/*
			key := r.Header.Get(NAUTHORIZATION)
			if appid, isValid := isValidKey(key); isValid {
				r.Header.Set(NAPPID, appid)
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			}
		*/
	})
}

func isValidKey(key string) (string, bool) {
	if key == "fail" {
		return "", false
	}
	return "fakeappid", true
}
