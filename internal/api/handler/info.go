package handler

import "net/http"

var body = `<!doctype html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Накопительная система лояльности «Гофермарт»</title>
	</head>
	<body>
		<h1>Накопительная система лояльности «Гофермарт»</h1>
		<p>
			Данный сервис предоставляет API для накопительная системы лояльности.
		</p>		
	</body>
</html>
`

func InfoPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(body))
}
