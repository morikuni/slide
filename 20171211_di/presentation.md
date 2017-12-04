theme: Plain Jane, 3
autoscale: true

# DIコンテナを使わないDI
### golang.tokyo#11 2017/12/11

---

# 自己紹介

![40%, right](morikuni.jpg)

- Name: 森國 泰平 (Morikuni Taihei)
- GitHub: [@morikuni](https://github.com/morikuni)
- Twitter: [@inukirom](https://twitter.com/inukirom)
- 所属
    - 株式会社メルカリ/ソウゾウ
    - メルカリ カウル
![inline, 10%](kauru_icon.png)

---

# 本日の内容

### ❌️ DIできるようにコードを書く方法
### ⭕️ DIコンテナを使わずに依存関係を解決する方法

---

# どうやって依存関係を解決する？

- mainに書く
- DIコンテナを使う
- DI用の関数を定義する

---

# 例題

以下の3つの関数を使って`Service`のインスタンスを作れ。

```go
func NewService(repo Repository, mailer Mailer) Service { ... }

func NewRepository(db *sql.DB) Repository { ... }

func NewMailer() Mailer { ... }
```

`Service`は`Repository`と`Mailer`に依存している。
`Repository`は`*sql.DB`に依存している。

---

# mainに書く


```go
func main() {
    db, err := sql.Open("db", "dsn")
    if err != nil { ... }
    repo := NewRepository(db)
    mailer := NewMailer()
    service := NewService(repo, mailer)
}
```

### オブジェクトが増える度にmainが肥大化していき可読性が低い

---

# DIコンテナを使う (goldi[^*1])

- yamlに依存関係を記述する
- DIコンテナ用のコードを生成する
- DIコンテナから必要なインスタンスを取得する

[^*1]: 実際にgoldiを使ったことはないので間違っている可能性があります

### DIコンテナの使い方を覚える必要がある

---

```yaml
types:
    db:
        package: database/sql
        type: *DB
        factory: Open
        arguments:
            - "db"
            - "dsn"
    repository:
        package: github.com/morikuni/hoge
        type: Repository
        factory: NewRepository
        arguments:
            - "@db"
    mailer:
        package: github.com/morikuni/hoge
        type: Mailer
        factory: NewMailer
    service:
        package: github.com/morikuni/hoge
        type: Service
        factory: NewService
        arguments:
            - "@repository"
            - "@mailer"
```

---

# DI用の関数を定義する(Inject関数)

- オブジェクトが引数0個で取得できるようにする関数
- あるオブジェクトについて、依存先が引数0個で取得できればそのオブジェクトも引数0個で取得できる
- Inject関数を組み合わせることでInject関数を作る

---

# Inject関数の実装

`InjectDB`は`*sql.DB`が引数0個で取得できるようにする

```go
func InjectDB() *sql.DB {
    db, err := sql.Open("db", "dsn")
    if err != nil {
        panic(err)
    }
    return db
}
```

---

# Inject関数の実装

`*sql.DB`が引数0個で取得できるので`Repository`も引数0個で取得できる

```go
func InjectRepository() Repository {
    return NewRepository(
        InjectDB(),
    )
}

func InjectMailer() Mailer {
    return NewMailer()
}
```

---

# Inject関数の実装

`Repository`と`Mailer`が引数0個で取得できるので、`Service`も引数0個で取得できるようになる

```go
func InjectService() Service {
    return NewService(
        InjectRepository(),
        InjectMailer(),
    )
}
```

---

# Inject関数を使う利点

- 依存先が増減したとしても影響範囲はInject関数内に収まる
- 実装が書かれるpackageが変わっても影響範囲はInject関数内に収まる
- Goのコードで書けるので構文を新しく覚える必要がない

---

# 詳しくはWebで！

### 明日(12/12)のQiita Advent Calender Go4に
### より詳しい解説記事を書きます！

#### [https://qiita.com/advent-calendar/2017/go4](https://qiita.com/advent-calendar/2017/go4)

