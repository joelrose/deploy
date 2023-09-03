#!/bin/sh

# Start the Docker daemon
dockerd &
DOCKER_PID=$!

# Start the SSH daemon
/usr/sbin/sshd -D &
SSHD_PID=$!

# Trap when this script ends and kill both processes
trap "kill $DOCKER_PID $SSHD_PID" EXIT

# Wait indefinitely
while true; do
    sleep 60
done
