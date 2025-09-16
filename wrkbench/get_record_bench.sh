#!/bin/bash

wrk -t8 -c100 -d5s -T1s --script=get_record.lua --latency  "http://127.0.0.1:8109/msg/get_msg_record"
