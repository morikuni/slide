theme: Plain Jane, 3
autoscale: true

# [fit]強いサービスの作り方
# @morikuni

---

# 強いサービスとは

---

# リアクティブ宣言

>求められているのは、システムアーキテクチャに対する明快なアプローチであると我々は考える。そして、必要な側面の全ては既に独立に認識されている: 求めるものは、即応性と、耐障害性と、弾力性と、メッセージ駆動とを備えたシステムだ。我々はこれをリアクティブシステム (Reactive Systems) と呼ぶ。
-- [http://www.reactivemanifesto.org/ja](http://www.reactivemanifesto.org/ja)

---

## リアクティブシステムとは

- 即応性 (Responsive)
  - システムは可能な限りすみやかに応答する。
- 耐障害性 (Resilient)
  - システムは障害に直面しても即応性を保ち続ける。
- 弾力性 (Elastic)
  - システムはワークロードが変動しても即応性を保ち続ける。
- メッセージ駆動 (Message Driven)
  - リアクティブシステムは非同期なメッセージパッシングに依ってコンポーネント間の境界を確立する。

---

# 本日はアプリケーションレベルの
# 耐障害性についてお話しします

---

# 耐障害性のあるサービス

- 死なない
- 外部リソースが死んでいても重くならない
- (外部リソースを殺さない)

---

# 耐障害性を持たせるためのパターン

- Timeout
- Retry
- Rate Limit
- Bulkhead
- Circuit Breaker
- and more...

---

# Timeout

|||
|:-:|:-:|
|障害|高負荷などにより処理が低速になる|
|障害の影響|低速な処理を行うサービスが低速になる。そのサービスにアクセスするサービスが低速になる。そのサービス…|
|対策|時間がかかる可能性のある処理に制限時間を設ける|
|うれしいところ|低速な外部リソースの影響を受けない|
|Goでの実装|`context.WithTimeout`, `context.WithDeadline`|
|||

---

# Timeout

```go
ctx, cancel := context.WithTimeout(ctx, time.Second)
defer cancel()

select {
case <-ctx.Done():
    return ctx.Err()
case result := <-Process(ctx):
    return result
}
```

---

# Retry

|||
|:-:|:-:|
|障害|ネットワークが切断される。外部リソースのプロセスが再起動する。etc...|
|障害の影響|処理が失敗する|
|対策|処理が失敗した場合に適切か間隔をあけて再試行する|
|うれしいところ|一時的な障害から自動で回復する|
|Goでの実装|`github.com/Songmu/retry`|
|||

---

# Retry

```go
err := retry.Retry(3, time.Second, func() error {
    return Process()
})

func Retry(n uint, interval time.Duration, fn func() error) (err error) {
	for n > 0 {
		n--
		err = fn()
		if err == nil || n <= 0 {
			break
		}
		time.Sleep(interval)
	}
	return err
}
```

---

# Rate Limit

|||
|:-:|:-:|
|障害|想定以上に処理が実行され高負荷になる|
|障害の影響|処理が失敗する。低速になる。etc...|
|対策|特定の期間に処理が実行できる回数を制限する|
|うれしいところ|負荷を制御することが出来る|
|Goでの実装|`golang.org/x/time/rate`|
|||

---

# Rate Limit

```go
limiter := rate.NewLimiter(rate.Every(time.Secont/1000, 5000))
err := limiter.Wait(ctx)
if err != nil {
    return err
}
Process()
```

---

# Bulkhead

|||
|:-:|:-:|
|障害|一部の処理がCPUを使いすぎる|
|障害の影響|サービス全体が高負荷になる|
|対策|各処理が使える計算リソースを分離する|
|うれしいところ|負荷を制御することが出来る|
|Goでの実装|`https://github.com/Jeffail/tunny` goroutineのスケジューラーは触れないので厳密に分離はできない|
|||

---

# Bulkhead

```go
pool, _ := tunny.CreatePool(10, func(object interface{}) interface{} {
    return Process(object.(string))
}).Open()

defer pool.Close()

result, err := pool.SendWork("hello")
```

---

# Circuit Breaker

|||
|:-:|:-:|
|障害|外部リソースが高負荷になり処理に失敗する|
|障害の影響|高負荷なところにアクセスを続けるので負荷が下がらない|
|対策|外部リソースの障害を検知したら以降はアクセスしない|
|うれしいところ|外部リソースが復旧できる可能性が上がる。無駄なアクセスを減らせるので高速になる。|
|Goでの実装|`https://github.com/rubyist/circuitbreaker`|
|||

---

# Circuit Breaker

```go
cb := circuit.NewThresholdBreaker(10)
err := cb.Call(func() error {
    return Process()
}, timeout * time.Seconds)
```

