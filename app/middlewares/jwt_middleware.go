package middlewares

//
//func JWTMiddleware(next http.Handler) http.Handler {
//
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		// Do stuff
//		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
//		if len(authHeader) != 2 {
//
//			fmt.Println("Malformed token")
//			w.WriteHeader(http.StatusUnauthorized)
//			w.Write([]byte("Malformed Token"))
//
//		} else {
//
//			jwtToken := authHeader[1]
//			token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
//				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
//					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
//				}
//				SECRETKEY := "secret"
//				return []byte(SECRETKEY), nil
//			})
//			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
//
//				ctx := context.WithValue(r.Context(), "props", claims)
//				next.ServeHTTP(w, r.WithContext(ctx))
//
//			} else {
//
//				fmt.Println(err)
//				w.WriteHeader(http.StatusUnauthorized)
//				w.Write([]byte("Unauthorized"))
//
//			}
//
//		}
//
//	})
//
//}
