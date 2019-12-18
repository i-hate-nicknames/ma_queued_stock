#!/bin/sh
printf "Creating order [1, 2, 3]\n"
http POST :8001/submit items:="[1, 2, 3]"

printf "Getting status\n"
http POST :8001/getStatus orderId:="1"

printf "Cancelling\n"
http POST :8001/cancel orderId:="1"