#!/usr/bin/expect
#!/bin/bash
set timeout 30
spawn ./gptn account new
expect "Passphrase:"
send "palletone@!@#$%^\r"
expect "Repeat passphrase:"
send "palletone@!@#$%^\r"  
interact

