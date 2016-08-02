RELEASE=RELEASE_381 # this may have to be changed based on llvm version
svn co https://llvm.org/svn/llvm-project/llvm/tags/$RELEASE/final $GOPATH/src/llvm.org/llvm
cd $GOPATH/src/llvm.org/llvm/bindings/go
./build.sh
go install llvm.org/llvm/bindings/go/llvm
