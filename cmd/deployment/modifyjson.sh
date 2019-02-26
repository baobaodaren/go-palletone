#!/bin/bash

function ModifyJson()
{

filename=../node1/ptn-genesis.json

index=$[ $4 - 1 ]

add=`cat $filename | jq ".initialMediatorCandidates[$index] |= . + {\"account\": \"$1\", \"initPubKey\": \"$2\", \"node\": \"$3\"}"`

if [ $index -eq 0 ] ; then

    createaccount=`./createaccount.sh`
    account=`echo $createaccount | sed -n '$p'| awk '{print $NF}'`
    account=${account:0:35}
    account=`echo ${account///}`

    add=`echo $add |
       jq "to_entries |
       map(if .key == \"tokenHolder\"
          then . + {\"value\":\"$account\"}
          else .
          end
         ) |
      from_entries"`

    createaccount=`./createaccount.sh`
    account=`echo $createaccount | sed -n '$p'| awk '{print $NF}'`
    account=${account:0:35}
    account=`echo ${account//^M/}`

    add=`echo $add |
       jq "to_entries |
       map(if .key == \"foundationAddress\"
          then . + {\"value\":\"$account\"}
          else .
          end
         ) |
      from_entries"`

fi

    rm $filename
    echo $add >> temp.json
    jq -r . temp.json >> $filename
    rm temp.json
}

