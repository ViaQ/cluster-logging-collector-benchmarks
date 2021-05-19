#!/bin/bash
export PATH=/usr/bin:$PATH
PAYLOAD_GEN=${PAYLOAD_GEN:-fixed}
DISTRIBUTION=${DISTRIBUTION:-fixed}
PAYLOAD_GEN=${PAYLOAD_GEN:-constant}
PAYLOAD_SIZE=${PAYLOAD_SIZE:-100}
MSG_PER_SEC=${MSG_PER_SEC:-1}
OUTPUT_FORMAT=${OUTPUT_FORMAT:-stdout}
OUT_FILE=${OUT_FILE:-""}
TOT_MSG=${TOT_MSG:-1}

log-stressor \
    -payload-gen   $PAYLOAD_GEN \
    -distribution  $DISTRIBUTION \
    -payload-gen   $PAYLOAD_GEN \
    -payload_size  $PAYLOAD_SIZE \
    -msgpersec     $MSG_PER_SEC \
    -output-format $OUTPUT_FORMAT \
    -totMessages   $TOT_MSG \
    -file          "$OUT_FILE"
