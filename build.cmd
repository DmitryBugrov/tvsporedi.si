PWD=`pwd`
GOPATH=%PWD%:%GOPATH
echo %GOPATH%

go build ./src/spored.tv.go

