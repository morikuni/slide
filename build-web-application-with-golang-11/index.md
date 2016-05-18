class: center, middle

# エラー処理、デバッグとテスト

[https://github.com/astaxie/build-web-application-with-golang/blob/master/ja/11.0.md](https://github.com/astaxie/build-web-application-with-golang/blob/master/ja/11.0.md)

---

## はじめに

多くのプログラマはbugの調査と修正の多くの時間をかけている。  
エラー処理に時間をかけたくはないけど、Webアプリケーションにエラーは付きのもで避けられない。  

Goでどのようにエラー処理を行うか、GDBによるデバッグの方法、Goのユニットテストについて紹介する。

---

## Cのエラー処理

Cでは`-1`や`NULL`でエラーを表している。  
でも、ユーザーはAPIのドキュメントを読まないとこれらの値が何を意味するのかわからない。

```c
FILE *fp = fopen(filepath, "r"); // FILE* or NULL
if (fp == NULL) {
	printf("%d\n", strerror(errno)); // グローバル変数errnoにエラーの詳細が入っている
}
```

---

## Javaのエラー処理

Javaでは例外によってエラーを表す。

```java
try {
	FileInputStream in
		= new FileInputStream("hoge.txt");
} catch (IOException e) {
	e.printStackTrace();
} finally {
	in.close();
}
```

---

## Goのエラー処理

`error`型と`nil`によってエラーを表す。  
エラーが発生し得る関数は`error`を返すように設計されている。

```go
f, err := os.Open("hoge.txt")
if err != nil {
    log.Fatal(err)
}
```

---

## error型

`error`型はビルトインのインターフェース。
`Error()`メソッドを呼び出すことでエラーを表す文字列を取得できる。

```go
type error interface {
    Error() string
}
```

多くのエラーの実態は`errors`パッケージ内の構造体。

```go
type errorString struct {
    s string
}

func (e *errorString) Error() string {
    return e.s
}
```

---

## カスタム定義のエラー

`json`パッケージのエラー。

```go
type SyntaxError struct {
    msg    string // エラーの説明
    Offset int64  // エラーが発生した場所
}

func (e *SyntaxError) Error() string { return e.msg }
```

```go
if err := dec.Decode(&val); err != nil {
    if serr, ok := err.(*json.SyntaxError); ok {
        line, col := findLine(f, serr.Offset)
        return fmt.Errorf("%s:%d:%d: %v", f.Name(), line, col, err)
    }
    return err
}
```

---

## 注意?

>関数がカスタム定義のエラーを返す時は戻り値にerror型を設定するようおすすめします。  
>特に前もって宣言しておく必要のないエラー型の変数には注意が必要です。

```go
func Decode() *SyntaxError {
    var err *SyntaxErro // 予めエラー変数を宣言します
    if エラー条件 {
        err = &SyntaxError{}
    }
    return err // errは永久にnilではない値と等しくなる
}
```

え？`nil`になったけど・・・

---

## Webアプリケーションのエラー

毎回`http.Error`するとエラー処理のロジックが多くなる。

```go
func init() {
    http.HandleFunc("/view", viewRecord)
}

func viewRecord(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    key := datastore.NewKey(c, "Record", r.FormValue("id"), 0, nil)
    record := new(Record)
    if err := datastore.Get(c, key, record); err != nil {
*        http.Error(w, err.Error(), 500)
        return
    }
    if err := viewTemplate.Execute(w, record); err != nil {
*        http.Error(w, err.Error(), 500)
    }
}
```

---

### エラーを返せるハンドラを定義する

```go
type appHandler func(http.ResponseWriter, *http.Request) error

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
```

```go
func init() {
    http.Handle("/view", appHandler(viewRecord))
}

func viewRecord(w http.ResponseWriter, r *http.Request) error {
    c := appengine.NewContext(r)
    key := datastore.NewKey(c, "Record", r.FormValue("id"), 0, nil)
    record := new(Record)
    if err := datastore.Get(c, key, record); err != nil {
*        return err
    }
    return viewTemplate.Execute(w, record)
}
```

---

### エラーコードを指定できるようにする

```go
type appError struct {
    Error   error
    Message string
    Code    int
}
```

```go
type appHandler func(http.ResponseWriter, *http.Request) *appError

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if e := fn(w, r); e != nil { // e is *appError, not os.Error.
        c := appengine.NewContext(r)
        c.Errorf("%v", e.Error)
        http.Error(w, e.Message, e.Code)
    }
}
```

---

```go
func viewRecord(w http.ResponseWriter, r *http.Request) *appError {
    c := appengine.NewContext(r)
    key := datastore.NewKey(c, "Record", r.FormValue("id"), 0, nil)
    record := new(Record)
    if err := datastore.Get(c, key, record); err != nil {
*        return &appError{err, "Record not found", 404}
    }
    if err := viewTemplate.Execute(w, record); err != nil {
*        return &appError{err, "Can't display record", 500}
    }
    return nil
}
```

---

## GDBを使用してデバッグする

GDBでは以下のようなことができる。

- プログラムを起動し、開発者の要求に従ってプログラムを実行する
- 開発者が設定したブレークポイントで実行を停止する
- プログラムが停止したとき、プログラムの状態を調べる
- 動的にプログラムの実行環境を変更する

詳細は割愛。

---

## Goのデバッグツール

- fmt.printf
- GDB
- godebug
- Delve
	- リモートデバッグができるのでエディタのプラグインとかにつかわれている

---

## Goでどのようにテストを書くか

`testing`パッケージと`go test`コマンドでテストを書くことができる。

- テストコードのファイル名は`_test.go`で終わる
- テスト関数の名前は`Test`から始まる
- 関数名で`Test`に続く文字列は小文字で始まってはいけない
- テストは上から順に実行される
- テスト関数の引数は`testing.T`
- `Error`, `Errorf`, `FailNow` `Fatal`, `Fatalf`でテストを失敗させる
- `Log`で情報を出力

---

- `go test`でテストを実行

```profile
--- FAIL: Test_Division_2 (0.00 seconds)
    gotest_test.go:16: パスしません
FAIL
exit status 1
FAIL    gotest  0.013s
```

- `go test -v`で実行された全てのテストの情報を出力

```profile
=== RUN Test_Division_1
--- PASS: Test_Division_1 (0.00 seconds)
    gotest_test.go:11: 1つ目のテストがパス
=== RUN Test_Division_2
--- FAIL: Test_Division_2 (0.00 seconds)
    gotest_test.go:16: パスしません
FAIL
exit status 1
FAIL    gotest  0.012s
```

---

## どのように耐久テスト(ベンチマーク)を書くか

- 基本的にはテストの書き方と同じ
- テスト関数の名前は`Benchmark`から始まる
- テスト関数の引数は`testing.B`
- テスト中のループは`testing.B.N`を使う

---

- `go test -bench`で実行

```profile
PASS
BenchmarkHoge-4	2000000000	         0.31 ns/op
ok  	_/Users/morikuni-taihei/Repositories/slide/build-web-application-with-golang-11	0.658s
```

- `go test -bench -benchmem`

```profile
PASS
BenchmarkHoge-4	2000000000	         0.31 ns/op	       0 B/op	       0 allocs/op
ok  	_/Users/morikuni-taihei/Repositories/slide/build-web-application-with-golang-11	0.652s
```

---

## まとめ

テストを書くことで将来コードが変更されたときにも回帰テストを行うことができる。  
ユニットテストと耐久テストは実運用が開始された後のコードが予定通り実行されることを保証してくれる。

---

## おまけ

[ユニットテストが書きやすい設計〜自家製モックを添えて〜 - Qiita](http://qiita.com/morikuni/items/24c98e5d8116ab14fcdf)
