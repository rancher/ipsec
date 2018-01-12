#!/bin/bash
set -e -x

trap "exit 1" SIGTERM SIGINT

while curl http://localhost:8111 >/dev/null 2>&1; do
    # This is an upgrade hack from going from v0.7.5 to something newer
    echo Waiting for old ipsec container to stop
    sleep 2
done

export CHARON_PID_FILE=/var/run/charon.pid
rm -f ${CHARON_PID_FILE}

GCM=false

for ((i=0; i<6; i++)); do
    if ip xfrm state add src 1.1.1.1 dst 1.1.1.1 spi 42 proto esp mode tunnel aead "rfc4106(gcm(aes))" 0x0000000000000000000000000000000000000001 128 sel src 1.1.1.1 dst 1.1.1.1; then
        GCM=true
        ip xfrm state del src 1.1.1.1 dst 1.1.1.1 spi 42 proto esp 2>/dev/null || true
        break
    fi
    ip xfrm state del src 1.1.1.1 dst 1.1.1.1 spi 42 proto esp 2>/dev/null || true
    sleep 1
done

if [ "${RANCHER_DEBUG}" == "true" ]; then
    DEBUG="--debug"
else
    DEBUG=""
fi


mkdir -p /etc/ipsec
if [ "${RANCHER_IPSEC_PSK}" != "" ]; then
    echo "${RANCHER_IPSEC_PSK}" > /etc/ipsec/psk.txt
else
    if [ "$TOKEN_PSK" = "true" ]; then
        curl -s -f http://169.254.169.250/2016-07-29/self/service/token > /etc/ipsec/psk.txt
        if [ -z /etc/ipsec/psk.txt ]; then
            echo Failed to download PSK
            exit 1
        fi
    else
        curl -f -u ${CATTLE_ACCESS_KEY}:${CATTLE_SECRET_KEY} ${CATTLE_URL}/configcontent/psk > /etc/ipsec/psk.txt
        curl -f -X PUT -d "" -u ${CATTLE_ACCESS_KEY}:${CATTLE_SECRET_KEY} ${CATTLE_URL}/configcontent/psk?version=latest
    fi
fi

GATEWAY=$(ip route get 8.8.8.8 | awk '{print $3}')
iptables -t nat -I POSTROUTING -o eth0 -s $GATEWAY -j MASQUERADE
exec rancher-ipsec \
--gcm=$GCM \
--charon-launch \
--ipsec-config /etc/ipsec \
${DEBUG}
