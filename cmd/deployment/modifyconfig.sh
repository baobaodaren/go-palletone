#!/bin/bash
source ./modifyjson.sh



HTTPPort=8545
WSPort=8546
Port=8080
ListenAddr=30303
BtcHost=18332
ContractAddress=12345
LocalHost=localhost



function ModifyConfig()
{

dumcpconfig=`./gptn dumpconfig`
echo $dumpconfig

if [ $1 -ne 1 ] ;then
newipcpath="IPCPath=\"gptn$1.ipc\""
sed -i '/^IPCPath/c'$newipcpath'' ptn-config.toml

newHTTPPort="HTTPPort=$[$HTTPPort+$1*10]"
sed -i '/^HTTPPort/c'$newHTTPPort'' ptn-config.toml

newWSPort="WSPort=$[$WSPort+$1*10]"
sed -i '/^WSPort/c'$newWSPort'' ptn-config.toml

newPort="Port=$[$Port+$1]"
sed -i '/^Port/c'$newPort'' ptn-config.toml


newListenAddr="ListenAddr=\":$[$ListenAddr+$1]\""
sed -i '/^ListenAddr/c'$newListenAddr'' ptn-config.toml

newBtcHost="BtcHost=\"localhost:$[$BtcHost+$1]\""
sed -i '/^BtcHost/c'$newBtcHost'' ptn-config.toml


newContractAddress="ContractAddress=\"127.0.0.1:$[$ContractAddress+$1]\""
sed -i '/^ContractAddress/c'$newContractAddress'' ptn-config.toml
else
dumcpjson=`./gptn dumpjson`
echo $dumpjson

newEnableStaleProduction="EnableStaleProduction=true"
sed -i '/^EnableStaleProduction/c'$newEnableStaleProduction'' ptn-config.toml

fi


createaccount=`./createaccount.sh`
tempinfo=`echo $createaccount | sed -n '$p'| awk '{print $NF}'`
accountlength=35
accounttemp=${tempinfo:0:$accountlength}
#account=`echo ${accounttemp//^M/}`
account=`echo ${accounttemp///}`




newAddress="Address=\"$account\""
sed -i '/^Address/c'$newAddress'' ptn-config.toml


newPassword="Password=\"palletone@!@#$%^\""
sed -i '/^Password/c'$newPassword'' ptn-config.toml




info=`./gptn mediator initdks`
key=`echo $info`

privatekeylength=44
private=${key#*private key: }
privatekeytemp=${private:0:$privatekeylength}
privatekey=`echo ${privatekeytemp///}`
#echo $privatekey


publickeylength=175
public=${key#*public key: }
publickeytemp=${public:0:$publickeylength}
publickey=`echo ${publickeytemp///}`
#echo $publickey


newInitPrivKey="InitPrivKey=\"$privatekey\""
sed -i '/^InitPrivKey/c'$newInitPrivKey'' ptn-config.toml


newInitPubKey="InitPubKey=\"$publickey\""
sed -i '/^InitPubKey/c'$newInitPubKey'' ptn-config.toml



while :
do
info=`./gptn nodeInfo`
tempinfo=`echo $info | sed -n '$p'| awk '{print $NF}'`
length=`echo ${#tempinfo}`
nodeinfotemp=${tempinfo:0:$length}
nodeinfo=`echo ${nodeinfotemp//^M/}`
length=`echo ${#nodeinfo}`
b=140
if [ "$length" -lt "$b" ]
then
    continue
else
    break
fi


done



echo "account: "$account
echo "publickey: "$publickey
echo "nodeinfo: "$nodeinfo

ModifyJson  $account $publickey $nodeinfo
}


function addBootstrapNodes()
{
    filename=node1/ptn-genesis.json
    nodes=$1
    index=$2
    content=`cat $filename`
    #echo "node number:"$nodes
    #echo "index:"$index
    acount=1
    array="["
    while [ $acount -le $nodes ] ;
    do
	if [ $acount -ne $index ];then
	    #echo $acount
	    nodeinfo=`echo $content | jq ".initialMediatorCandidates[ $[$acount-1] ].node"`
	    array="$array$nodeinfo,"
        fi
	let ++acount;
    done
    l=${#array}
    if [ $l -eq 1 ] ;then
        newarr="[]"
    else
        newarr=${array:0:$[$l-1]}
        newarr="$newarr]"
    fi
    newBootstrapNodes="BootstrapNodes=$newarr"
    #sed -i '/^StaticNodes/c'$newStaticNodes'' node$index/ptn-config.toml
    sed -i '/^BootstrapNodes/c'$newBootstrapNodes'' node$index/ptn-config.toml
    echo "=====addBootstrapNodes $index ok======="
}




function ModifyBootstrapNodes()
{
    count=1;
    while [ $count -le $1 ] ;
    do
	#echo $count
        addBootstrapNodes $1 $count
        let ++count;
        sleep 1;
    done
    find . -name "*.toml" | xargs sed -i -e "s%\[\:\:\]%127.0.0.1%g"
    return 0;
}


