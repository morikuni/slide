class: center, middle

#Cottonの紹介

---

## 自己紹介

- 森國泰平(モリクニ タイヘイ)
- 16新卒 技術部
- 2月下旬からインターン、4月にそのまま入社
- GitHub@[morikuni](https://github.com/morikuni)
- Twitter@[inukirom](https://twitter.com/inukirom)

---

class: center, middle

###CottonというGoのライブラリを作り始めたので紹介します

https://github.com/morikuni/cotton

---

## Cotton

_Simple, Lightweight and Composable HTTP Handler/Middleware_

- 標準の`net/http`と一緒に使うことを想定
- `http.HandlerFunc`を拡張するミドルウェアを作れる
- 柔軟なミドルウェアの合成方法を提供
- 成功時とエラー時の処理を分離、エラー処理を1箇所にかける

---

## 4つの型

```go
type http.HandlerFunc
  func(http.ResponseWriter, *http.Request)

type Middleware
  func(http.ResponseWriter, *http.Request, Service) Error

type Service
  func(http.ResponseWriter, *http.Request) Error

type ErrorHandler
  func(http.ResponseWriter, *http.Request, Error)
```

```txt
Middleware + Middleware       => Middleware
Middleware + http.HandlerFunc => Service
Middleware + Service          => Service
Service    + ErrorHandler     => http.HandlerFunc
```

---

## 考え方

http://niconare.nicovideo.jp/watch/kn1052#9

---

## サンプル

https://github.com/morikuni/cotton#example

---

## ミドルウェアの作り方

```go
import "net/http"

func MyMiddleware(w http.ResponseWriter, r *http.Request, s Service) Error {
	err := DoSomething()
	if err != nil {
		return err
	}
	err2 := s(w, r)
	err3 := DoSomething2(err2)
	return err3;
}
```

[PanicFilter](https://github.com/morikuni/cotton/blob/master/panic.go)

[MethodFilter](https://github.com/morikuni/cotton/blob/master/method.go)

---

## 今後

ミドルウェアを追加していきたい



