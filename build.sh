WORKROOT=$(pwd)
export GOPATH=${WORKROOT}

# try to download bce-sdk-go from github
go env -w GO111MODULE=on
if [ $? -ne 0 ]
then
    echo "fail to set env GO111MODULE"
    exit 1
fi

go env -w GONOPROXY=\*\*.baidu.com\*\*
go env -w GOPROXY=https://goproxy.baidu.com
go env -w GONOSUMDB=\*

chmod +w pkg -R 2>&1 > /dev/null
rm -rf pkg/mod/github.com/baidubce/bce-sdk-go* 
rm -rf ./src/github.com/baidubce/bce-sdk-go

go get -d github.com/baidubce/bce-sdk-go@1a69080
if [ $? -ne 0 ]
then
    echo "fail to get bce-sdk-go"
    exit 1
fi

chmod +w ./pkg -R
mkdir -p ./src/github.com/baidubce/
mv pkg/mod/github.com/baidubce/bce-sdk-go* ./src/github.com/baidubce/bce-sdk-go

go env -w GO111MODULE=off

# start to build bcecmd
cd $WORKROOT/src/main

go build bcecmd.go
if [ $? -ne 0 ];
then
	echo "fail to build bcecmd.go"
	exit 1
fi

cd ../../
if [ -d "./output" ]
then
	rm -rf output
fi

mkdir output
mv src/main/bcecmd "output/bcecmd"
chmod +x output/bcecmd 

echo "OK for build bcecmd"
