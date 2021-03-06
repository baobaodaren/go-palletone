mkdir -p $GOPATH/src/github.com/golang
cd $GOPATH/src/github.com/golang
git clone https://github.com/golang/mock.git
cd mock
git checkout 442550a
go install github.com/golang/mock/mockgen
cd $GOPATH/src/github.com/palletone/go-palletone
mockgen -source=./dag/interface.go -destination=./dag/dag_mock.go -package=dag -self_package="github.com/palletone/go-palletone/dag"
mockgen -source=./dag/txspool/interface.go -destination=./dag/txspool/txpool_mock.go -package=txspool -self_package="github.com/palletone/go-palletone/dag/txspool"
mockgen -source=./ptn/mediator_connection.go  -destination=./ptn/mediator_connection_mock.go -package=ptn
mockgen -source=./contracts/shim/interfaces_stable.go  -destination=./contracts/shim/interfaces_stable_mock.go -package=shim -self_package="github.com/palletone/go-palletone/contracts/shim"
