#!/bin/bash
# Simulates network latency, packet loss, and jitter using tc (traffic control)
# Usage: ./simulate_latency.sh <start|stop> [interface] [latency_ms] [packet_loss_percent] [jitter_ms]

ACTION=$1
INTERFACE=${2:-eth0}
LATENCY=${3:-100}
LOSS=${4:-2}
JITTER=${5:-20}

if [ "$ACTION" == "start" ]; then
    echo "Starting network simulation on $INTERFACE: ${LATENCY}ms latency, ${LOSS}% loss, ${JITTER}ms jitter"
    tc qdisc add dev $INTERFACE root handle 1: netem delay ${LATENCY}ms ${JITTER}ms distribution normal loss ${LOSS}%
elif [ "$ACTION" == "stop" ]; then
    echo "Stopping network simulation on $INTERFACE"
    tc qdisc del dev $INTERFACE root
else
    echo "Usage: $0 <start|stop> [interface] [latency_ms] [packet_loss_percent] [jitter_ms]"
    exit 1
fi
