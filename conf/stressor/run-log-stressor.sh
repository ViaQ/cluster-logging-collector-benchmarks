#!/bin/bash
export PATH=/usr/bin:$PATH
PAYLOAD_GEN=${PAYLOAD_GEN:-fixed}
DISTRIBUTION=${DISTRIBUTION:-fixed}
PAYLOAD_GEN=${PAYLOAD_GEN:-constant}
PAYLOAD_SIZE=${PAYLOAD_SIZE:-100}
MSG_PER_SEC=${MSG_PER_SEC:-100}
OUTPUT_FORMAT=${OUTPUT_FORMAT:-stdout}
OUT_FILE=${OUT_FILE:-}
TOT_MSG=${TOT_MSG:-1}

# ARGS="-payload-gen   $PAYLOAD_GEN \
#     -distribution  $DISTRIBUTION \
#     -payload-gen   $PAYLOAD_GEN \
#     -payload_size  $PAYLOAD_SIZE \
#     -msgpersec     $MSG_PER_SEC \
# if [ -n "$OUT_FILE" ] ; then
#     ARGS="$ARGS -file $OUT_FILE"
# fi
log-stressor $@