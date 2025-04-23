#!/bin/bash

set -e

export DEBIAN_FRONTEND=noninteractive
echo "Starting update job at $(TZ=Europe/Berlin date +%Y-%m-%d_%H:%M:%S)"
echo "Pruning idle docker images"
docker image prune -af
echo "Updating"
apt-get update
echo "Upgrading"
apt-get upgrade -y
echo "Removing no longer required packages"
apt-get autoremove --purge
echo "Rebooting"
shutdown -r now
echo ""
