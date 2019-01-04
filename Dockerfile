FROM golang:latest

RUN mkdir -p $GOPATH/src/github.com/nimrodshn/accountpooloperator

WORKDIR $GOPATH/src/github.com/nimrodshn/accountpooloperator

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh 
    
COPY . $GOPATH/src/github.com/nimrodshn/accountpooloperator

RUN dep ensure -v

RUN make

CMD ["./accountpooloperator"]
