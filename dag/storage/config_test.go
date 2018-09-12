package storage

import (
	"fmt"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag/modules"
	"log"
	"testing"
	"github.com/palletone/go-palletone/common/ptndb"
)

func MockStateMemDb() StateDb{
	db,_:=ptndb.NewMemDatabase()
	statedb:=NewStateDatabase(db)
	return statedb
}

func TestSaveAndGetConfig(t *testing.T) {
	//Dbconn := storage.ReNewDbConn("E:\\codes\\go\\src\\github.com\\palletone\\go-palletone\\cmd\\gptn\\gptn\\leveldb")
	//if Dbconn == nil {
	//	fmt.Println("Connect to db error.")
	//	return
	//}
	db:=MockStateMemDb()
	confs := []modules.PayloadMapStruct{}
	aid := modules.IDType16{}
	aid.SetBytes([]byte("1111111111111111222222222222222222"))
	st := modules.Asset{
		AssetId:  aid,
		UniqueId: aid,
		ChainId:  1,
	}
	confs = append(confs, modules.PayloadMapStruct{Key: "TestStruct", Value: modules.ToPayloadMapValueBytes(st)})
	confs = append(confs, modules.PayloadMapStruct{Key: "TestInt", Value: modules.ToPayloadMapValueBytes(uint32(10))})
	stateVersion := modules.StateVersion{
		Height: modules.ChainIndex{
			AssetID: aid,
			IsMain:  true,
			Index:   0,
		},
		TxIndex: 0,
	}
	log.Println(stateVersion)
	if err := db.SaveConfig(confs, &stateVersion); err != nil {
		log.Println(err)
	}


	data := db.GetConfig( []byte("MediatorCandidates"))
	var mList []core.MediatorInfo
	fmt.Println(data)
	if err := rlp.DecodeBytes(data, &mList); err != nil {
		log.Println("Check unit signature when get mediators list", "error", err.Error())
		return
	}
	// todo get ActiveMediators
	bNum := db.GetConfig( []byte("ActiveMediators"))
	var mNum uint16
	if err := rlp.DecodeBytes(bNum, &mNum); err != nil {
		log.Println("Check unit signature", "error", err.Error())
		return
	}
	if int(mNum) != len(mList) {
		log.Println("Check unit signature", "error", "mediators info error, pls update network")
		return
	}
	log.Println(">>>>>>>>> Pass >>>>>>>>>>.")
}
//
//func TestSaveStruct(t *testing.T) {
//	//Dbconn := storage.ReNewDbConn(dagconfig.DbPath)
//	//if Dbconn == nil {
//	//	fmt.Println("Connect to db error.")
//	//	return
//	//}
//	db:=MockStateMemDb()
//	aid := modules.IDType16{}
//	aid.SetBytes([]byte("1111111111111111222222222222222222"))
//	st := modules.Asset{
//		AssetId:  aid,
//		UniqueId: aid,
//		ChainId:  1,
//	}
//
//	if err := storage.Store(Dbconn, "TestStruct", st); err != nil {
//		t.Error(err.Error())
//	}
//}
//
//func TestReadStruct(t *testing.T) {
//	Dbconn := storage.ReNewDbConn(dagconfig.DbPath)
//	if Dbconn == nil {
//		fmt.Println("Connect to db error.")
//		return
//	}
//
//	data, err := storage.Get(Dbconn, []byte("TestStruct"))
//	if err != nil {
//		t.Error(err.Error())
//	}
//
//	var st modules.Asset
//	if err := rlp.DecodeBytes(data, &st); err != nil {
//		t.Error(err.Error())
//	}
//	log.Println("Data:", data)
//	log.Println(st)
//}
