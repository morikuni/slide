package main

import (
	"fmt"
)

type SyntaxError struct {
	msg    string // エラーの説明
	Offset int64  // エラーが発生した場所
}

func (e *SyntaxError) Error() string { return e.msg }

func Decode(b bool) *SyntaxError { // エラー、上のレイヤでコールした利用者によるerr!=nilの判断が永遠にtrueになります。
	var err *SyntaxError // 予めエラー変数を宣言します
	if b {
		err = &SyntaxError{}
	}
	return err // エラー、errは永久にnilではない値と等しくなり、上のレイヤでコールした利用者によるerr!=nilの判断が常にtrueとなります
}

func main() {
	fmt.Println(Decode(true) == nil)
	fmt.Println(Decode(false) == nil)
}
