Config file:
สามารถแก้ไข config ได้ในไฟล์ config/config.go ก่อน install
- OmisePublicKey หาได้ผ่าน account ของ ommise
-	OmiseSecretKey หาได้ผ่าน account ของ ommise
-	MaxGoRoutine   go routine สูงสุด



How to use:
```sh
$ cd $GOPATH/omise/go-tamboon
$ go install -v .

$ $GOPATH/bin/go-tamboon /path/to/test.csv

```