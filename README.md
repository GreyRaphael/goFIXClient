# FIX Client by go

- [FIX Client by go](#fix-client-by-go)
  - [How to use](#how-to-use)
  - [How to test](#how-to-test)

## How to use

run the program temporarily: `go run .`

build the program: `go build`

build with smaller binary size: `go build -ldflags="-s -w"`


## How to test

```bash
# 使用input/single.csv下单子
.\gofix.exe

# 使用input/zz500.csv下500笔单子
.\gofix.exe -i .\input\zz500.csv

# 使用input/zz500.csv下2000笔单子，4批
.\gofix.exe -i .\input\zz500.csv -n 4
```