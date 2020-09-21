#!/bin/bash

function setup_service(){
	PORT="$1"
	NAME="$2"
	COLOR="$3"
	curl --request PUT \
		--url http://localhost:3333/api/services \
		--header 'content-type: application/json' \
		--data "{\"port\":$PORT,\"name\":\"$NAME\",\"color\":\"#$COLOR\",\"notes\":\"\"}"
}

BACKEND=false
FRONTEND=false
SETUP=true
IMPORT=true
START=false


if $START; then
	pkill caronte
fi

if $BACKEND ; then
	go mod download && go build || exit -1
fi

if $FRONTEND ; then
	cd frontend && yarnpkg install && yarnpkg build || exit -2
	cd -
fi

if $START; then
	docker run -d -p 27017-27019:27017-27019 --name mongodb mongo:4 && sleep 3
	./caronte &
	sleep 2
fi

# setup 
if $SETUP ; then
	curl \
		--header "Content-Type: application/json"  \
		--request POST \
		--data '{"config": {"server_address": "10.10.1.1", "flag_regex": "flg[a-zA-Z0-9]{25}", "auth_required": false}, "accounts": {"usr1": "pwd1"}}' \
		http://localhost:3333/setup

	setup_service 8080  crashair        E53935
	setup_service 27017 aircnc          5E35B1
	setup_service 80    lostpropertyhub F9A825
	setup_service 5555  theone 	    F9A435
	#setup_service 3306  crashair        E53935
fi


# import pcaps
if $IMPORT ; then
	PCAP_DIR="~/pcaps"
	for PCAP in $PCAP_DIR/*.pcap ; do
		echo "[+] Uploading $PCAP" && \
		curl -F "file=@$PCAP" "http://localhost:3333/api/pcap/upload"
	done
fi

