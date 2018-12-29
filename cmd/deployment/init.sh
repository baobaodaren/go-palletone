#!/usr/bin/expect

set timeout 30
spawn ./gptn init
expect "Passphrase:"
send "palletone@!@#$%^\r"
interact











