class: center, middle

# Goと暗号化

---

## はじめに

[gipher](https://github.com/morikuni/gipher)というものを作っております

```sh
% cat test.json
{
    "aaa": "aaa",
    "bbb": 111,
    "ccc": {
        "ddd": 123.456789123456789123456789,
        "eee": "string:eee",
        "fff": "fff"
    }
}

% GIPHER_PASSWORD=hoge gipher encrypt --format json -f test.json --pattern ccc | jq
{
  "aaa": "aaa",
  "bbb": 111,
  "ccc": {
    "ddd": "9apGqccxwbtL2j8GQxQaIpDRQDb7zd+Mk0CHfj1wQd8sHGMo+6LATXY/ZH0=",
    "eee": "ob6PqQ56UUisxzUHMAvlCVFzoSTkgPLLnRW07+71XXa0",
    "fff": "GK79lAvbwxoxY0DS9GbIVl4+T5wDspgKkfA="
  }
}
```

---

```sh
% GIPHER_PASSWORD=hoge gipher decrypt --format json -f encrypted.json --pattern ccc | jq
{
  "aaa": "aaa",
  "bbb": 111,
  "ccc": {
    "ddd": 123.45678912345679,
    "eee": "string:eee",
    "fff": "fff"
  }
}
```

---

## 暗号化とは

詳しいことが知りたければ[暗号技術入門 第3版　秘密の国のアリス](https://www.amazon.co.jp/dp/B015643CPE)。

- ハッシュ化
    - 不可逆
    - sha1, sha256, md5, ...
    - 入力を予測不可能な値に変換
- 暗号化
    - 可逆
    - DES, Triple DES, AES, ...
    - 入力を**鍵**を用いて予測不可能な値に変換し、**鍵**を用いて元の入力を復元する
    - 公開鍵暗号方式と共通鍵暗号方式がある
        - 今回対象とするのは共通鍵暗号方式

---

### 共通鍵暗号方式の問題

- 鍵の配送方法
    - 鍵がないと復号できないので、鍵を送らなければならない
    - 鍵が漏れてしまった場合は暗号化している意味がない
    - 鍵を安全に送れるのであれば、暗号化なんかしなくても元のデータを安全に送れるはず

---

## AWS KMS (Key Management Service)

- 鍵の管理をしてくれるサービス
    - データキー (データを暗号化する鍵)
        - すぐに破棄する
        - ダウンロードできる
    - マスターキー (データキーを暗号化する鍵)
        - 永続化される
        - ダウンロードできない(ユーザーが見る手段がない)

---

### 暗号化の流れ
- KMSでデータキーと暗号化されたデータキーを生成する(ランダムな値とそれをマスターキーで暗号化したもの)
- データキーで入力を暗号化する
- 暗号化された入力 & 暗号化されたデータキーを送信する

### 復号の流れ
- KMSで暗号化されたデータキーを復号しデータキーを生成する
- データキーで暗号化された入力を復号する

KMSにテキストを投げてなにも考えずに暗号化、復号をすることもできる。

[10分でわかる！Key Management Serviceの仕組み #cmdevio](http://dev.classmethod.jp/cloud/aws/10minutes-kms/)

---

## 暗号化方式について


.half-left[
- ブロック暗号
    - 固定長bit列を暗号化する
    - DES, Triple DES, AES, ...
    - [Wikipedia](https://ja.wikipedia.org/wiki/%E3%83%96%E3%83%AD%E3%83%83%E3%82%AF%E6%9A%97%E5%8F%B7)
- ストリーム暗号
    - 任意長のbit列を暗号化する
    - 鍵をシードとして疑似乱数を生成し、それを用いて暗号化する
    - MUGI, RC4, ...
    - [Wikipedia](https://ja.wikipedia.org/wiki/%E3%82%B9%E3%83%88%E3%83%AA%E3%83%BC%E3%83%A0%E6%9A%97%E5%8F%B7)
]
.half-right[
- 暗号利用モード(ストリーム暗号の一種)
    - ブロック暗号を複数回適用することで任意のbit列を暗号化する
    - 使用するブロック暗号を入れ替えることが出来る
    - CBC, CTR, ...
    - [Wikipedia](https://ja.wikipedia.org/wiki/%E6%9A%97%E5%8F%B7%E5%88%A9%E7%94%A8%E3%83%A2%E3%83%BC%E3%83%89)

AESとCBC or AESとCTRがオススメされていた気がする
]


---

class: center, middle

# Goと暗号化

`crypto/*`パッケージをつかう

---

### ブロック暗号

`crypto/chpher.Block`

```go
type Block interface {
        // BlockSize returns the cipher's block size.
        BlockSize() int

        // Encrypt encrypts the first block in src into dst.
        // Dst and src may point at the same memory.
        Encrypt(dst, src []byte)

        // Decrypt decrypts the first block in src into dst.
        // Dst and src may point at the same memory.
        Decrypt(dst, src []byte)
}
```

- `crypto/aes`
- `crypto/des`
- `crypto/tea`

---

### ストリーム暗号 & 暗号利用モード

`crypto/chpher.Stream`

```go
type Stream interface {
        // XORKeyStream XORs each byte in the given slice with a byte from the
        // cipher's key stream. Dst and src may point to the same memory.
        // If len(dst) < len(src), XORKeyStream should panic. It is acceptable
        // to pass a dst bigger than src, and in that case, XORKeyStream will
        // only update dst[:len(src)] and will not touch the rest of dst.
        XORKeyStream(dst, src []byte)
}
```

- `crypto/rc4`
- `crypto/cipher`

---

### パスワードによる暗号化の例

[https://github.com/morikuni/gipher/password_cryptor.go](https://github.com/morikuni/gipher/blob/master/password_cryptor.go)

### KMSによる暗号化の例

[https://github.com/morikuni/gipher/aws_kms_cryptor.go](https://github.com/morikuni/gipher/blob/master/aws_kms_cryptor.go)

---

## aws-sdk-goの認証

各サービスを呼び出すためには`aws/client.ConfigProvider`を用意する必要がある。

```go
type ConfigProvider interface {
    ClientConfig(serviceName string, cfgs ...*aws.Config) Config
}
```

`ConfigProvider`には`aws/session.Session`を使用する。
`session.Session`は3種類の`config`から認証情報を読み込む。以下優先度順。

- `*aws.Config`
    - 引数に渡された認証情報
- `session.envConfig`
    - 環境変数から読み込んだ認証情報
- `session.sharedConfig`
    - `~/.aws/`から読み込んだ認証情報

---

### AssumeRole

`envConfig.EnableSharedConfig`が`false`の場合は`~/.aws/config`を読み込まない(AssumeRoleの設定は通常ここに書く)ので以下のどれかによって設定ファイルを読み込ませる。

- `AWS_SDK_LOAD_CONFIG=1`の環境変数を設定
- `session.NewSessionWithOptions`で`SharedConfigState: session.SharedConfigEnable`を指定する
- `~/.aws/config`の内容を`~/.aws/credentials`にコピーする(たぶん)

---

### Sessionで使われる環境変数

- `AWS_ACCESS_KEY_ID`: アクセスキー
- `AWS_SECRET_ACCESS_KEY`: シークレットキー
- `AWS_SESSION_TOKEN`: セッショントークン
- `AWS_REGION`: リージョン
- `AWS_PROFILE`: プロファイル
- `AWS_SDK_LOAD_CONFIG`: `~/.aws/config`を読み込むか
- `AWS_SHARED_CREDENTIALS_FILE`: credentialsファイルのパス
- `AWS_CONFIG_FILE`: configファイルのパス

---

## まとめ

- Goで暗号化は簡単
- KMS使うとなにも考えなくていい
- AssumeRoleしたかったらSessionを作るときに`SharedConfigState: session.SharedConfigEnable`せよ
