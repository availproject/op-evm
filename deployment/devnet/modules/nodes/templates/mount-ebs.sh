#!/usr/bin/env bash

EBS_DEVICE="/dev/nvme1n1"

mkdir "${workspace}" -p

while ! lsblk -no FSTYPE "$EBS_DEVICE"; do
  echo "Waiting for the ebs device to get attached..."
  sleep 10
done

# Format the EBS_DEVICE as ext4 if the EBS_DEVICE is empty and not formattet
if [ -z "$(lsblk -no FSTYPE "$EBS_DEVICE")" ]; then
  mkfs.ext4 "$EBS_DEVICE"
fi

mount "$EBS_DEVICE" "${workspace}" -t ext4

echo "EBS device mounted."
