package jury

import "github.com/palletone/go-palletone/dag/modules"

//type NewTxEvent struct {
//	Tx *modules.Transaction
//}

//install deploy invoke stop
type ContractExeEvent struct {
	Tx *modules.Transaction
}

type ContractSigEvent struct {
	Tx *modules.Transaction
}