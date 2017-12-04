アプリケーションコードを書いていると依存関係をどう管理するかが重要です。
そこでこの記事ではGoにおけるDIをテーマに、依存関係をどう管理していくかを書こうと思います。

# DIとは？

DIはDependency Injection(依存性の注入)の略称です。
ある処理に必要なオブジェクト(や関数)を外部から注入(指定)できるようにする実装パターンです。依存するオブジェクトを注入すること自体を指す場合もあります。

基本的には依存するオブジェクトをInterfaceとして定義し、Interfaceにのみ依存させることで、実装を入れ替えられるようにするというものです。

# なぜDIするか？

DIする利点として挙げられるのは主に次の２点です。

1. ユニットテストが書きやすくなる
2. オブジェクト間の結合度を下げやすくなる

一番大きいのはユニットテストが書きやすくなることです。
依存先がInterfaceになっていることで、その実装を入れ替えることができ、テスト時にモックを使ったユニットテストができるようになります。

また、Interfaceになっていることで、内部のフィールドにアクセスできなくなり、オブジェクト間の結合度が下げやすくなるという利点もあります。
ただし、無闇にGetterやSetterを追加するとこの利点は失われてしまうので注意が必要です。

# GoでのDI

では本題のGoでDIするためにはどうすればいいのかに移ります。
Goでは依存先をinterfaceとして定義し、structのフィールドとして持たせることによって、DIができるようになります。

例題として、次のような機能を考えてみましょう。

- メールアドレスとパスワードでユーザーを作成し保存する
- 成功したら登録完了メールを送信する

実装するためには次のようなコードを書くことになるでしょう。

```go
type SignUpService interface {
    SignUp(email, password string) error
}

type signUpService struct {}

func (s signUpService) SignUp(email, password string) error {
    // to be implemented
}
```

`SignUpService`は`User`を保存する処理と、メールを送信する処理に依存します。
なので`User`を保存する処理として`UserRepository`、メールを送信するための処理として`Mailer`をinterfaceとして定義します。

```go
type UserRepository interface {
    Save(u User) error
}

type Mailer interface {
    SendEmail(to, message) error
}
```

この2つのinterfaceを使って`SignUpService`を完成させましょう。

```go
type signUpService struct {
    repo   UserRepository
    mailer Mailer
}

func (s signUpService) SignUp(email, password string) error {
    u := NewUser(email, password)
    if err := s.repo.Save(u); err != nil {
        return err
    }
    return s.mailer.SendEmail(email, "登録完了")
}

func NewSignUpService(repo UserRepository, mailer Mailer) SignUpService {
    return signUpService{
        repo,
        mailer,
    }
}
```

これで`SignUpService`は依存先がinterfaceのみになったため、モックを使ったユニットテストが書けます。
ただし、アプリケーションを実行するときにはこれで終わりではありません。
`UserRepository`や`Mailer`の実装を取得し、`NewSignUpService`に渡さなくては`SignUpService`が使えません。
そこで、`UserRepository`と`Mailer`のコンストラクタが必要になります。具体的な内部実装は気にする必要はありません。

```go
func NewUserRepository(db DB) UserRepository {
    ...
}

func NewMailer() {
    ...
}
```

`UserRepository`は`DB`に依存しています。`DB`は`*sql.DB`のメソッドを定義したinterfaceだとしましょう。
`Mailer`は依存がないため引数なしで実装を取得できます。

次はこのコンストラクタを使い、どうやって依存関係を解決するかについて次の3つの方法を紹介します。

- mainに書く
- DIコンテナを使う
- DI用の関数を定義する

## mainに書く

1つめはmainで全ての依存関係を解決する方法です。

```go
func main() {
    db, err := sql.Open("db", "dsn")
    if err != nil {
        panic(err)
    }
    repo := NewUserRepository(db)
    mailer := NewMailer()
    service := NewSignUpService(repo, mailer)
    ...
}
```

シンプルですが、使用するオブジェクト数に比例してmainが肥大化していくという問題があります。

## DIコンテナを使う

2つ目はDIコンテナを使う方法です。
Javaなどでは一般的なんじゃないかと思いますが、Goでは使っているところを見たことがありません。
例として[goldi](https://github.com/fgrosse/goldi)を使いますが、私も使ったことはないので間違っているところがあるかもしれません。

goldiではyamlで依存関係を解決します。

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
        type: UserRepository
        factory: NewUserRepository
        arguments:
            - "@db"
    mailer:
        package: github.com/morikuni/hoge
        type: Mailer
        factory: NewMailer
    service:
        package: github.com/morikuni/hoge
        type: SignUpService
        factory: NewSignUpService
        arguments:
            - "@repository"
            - "@mailer"
```

goldigenというコマンドにこのyamlを渡すことで`RegisterTypes`という関数が生成されます。
DIコンテナにこの関数を適用することで依存関係が解決できるようになります。

```go
func main() {
    registry := goldi.NewTypeRegistry()
    RegisterTypes(registry)
    container := goldi.NewContainer(registry, nil)

    service := container.MustGet("service").(SignUpService)
    ...
}
```

DIコンテナを使うことでmainが肥大化していくことはなくなります。
ただし、DIコンテナの使い方を覚える必要があったり、`inteface{}`で取得したものをキャストする必要があったりします。

## DI用の関数を定義する

3つめはDI用の関数を用意する方法です。
この関数をInject関数と呼ぶことにします。
Inject関数は次のような関数です。

- オブジェクトを引数0個で取得できるようにする
- オブジェクトのコンストラクタに対して他のInject関数を使ってオブジェクトを注入する

あるオブジェクトについて、依存先が引数0個で取得できれば、そのオブジェクトも引数0個で取得できるので、これを組み合わせるというものです。

実際に例を見ていきましょう。

```go
func InjectDB() DB {
    db, err := sql.Open("db", "dsn")
    if err != nil {
        panic(err)
    }
    return db
}

func InjectUserRepository() UserRepository {
    return NewUserRepository(
        InjectDB(),
    )
}

func InjectMailer() Mailer {
    return NewMailer()
}

func InjectSignUpService() SignUpService {
    return NewSignUpService(
        InjectUserRepository(),
        InjectMailer(),
    )
}

func main() {
    service := InjectSignUpService()
}
```

最初に`InjectDB`を定義しています。
これは`sql.Open`を使って`*sql.DB`を返す関数です。
`InjectDB`を使うことで`DB`が引数0個で取得できるので、`UserRepository`も引数0個で取得できるようになります。
同様に`SignUpService`も`InjectRepository`と`InjectMailer`を使うことで引数0個で取得できます。
このようにInject関数を組み合わせること依存関係を解決するのがInject関数です。

1つ気になるのは、`InjectDB`内で`panic`を使っていることです。
Goの文化としてはできる限り`panic`を使わないことが望ましいと思いますが、Inject関数を使うのはmain関数内だけのはずです。
つまりInject関数はアプリケーションの初期化の時だけに呼ばれることになります。
`sql.Open`などに失敗すると言うことは初期化に失敗したということなので、おかしな状態で起動するよりは`panic`で起動に失敗してしまったほうがよいのではないかと思います。
どうしてもpanicさせたくなければ、ログを吐いた後で`os.Exit`するという方法でも構いません。

Inject関数を定義することの利点としては次のようなことが挙げられます。

- 依存先が増減しても影響範囲はInject関数内のみ
    - 例: `SignUpService`が`Logger`に依存するようになっても、`InjectSignUpService`に1行足すだけでよい
- 実装が書かれるpackageが変わっても影響範囲はInject関数内のみ
    - 例: `DB`の実装が`database/sql`パッケージから`datastore`パッケージに変わっても`InjectDB`で呼び出すコンストラクタを変更するだけでよい (interfaceが同じ限りは)

機能の追加やリファクタリングなどがやりやすくなるので、Inject関数を使う方法がよいのではないかと思います。

最後に、Inject関数をどのパッケージに書くかということですが、私はdiパッケージを作るのをオススメします。
専用のパッケージがあることによって、interfaceと実装の対応や、オブジェクトの注入など、依存関係に関するほとんどすべての責務を1つのパッケージに収めることができるからです。
また、diパッケージがmain以外から依存されないようにすることによって、Cycle Importも起きにくくなります。


# まとめ

- 依存するオブジェクトinterfaceにすることで実装を入れ替えることができ、ユニットテストが書きやすくなる
- Inject関数を定義することでDIコンテナを使わずに依存関係を解決することができ、リファクタリングなどもしやすくなる


