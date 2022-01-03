# Cluster Recover for RQLite running on a k8s cluster

The goal is to be able to recover a rqlite cluster when the majority of nodes get re-schedule to different nodes. It will basically generate a peers.json file everytime it detects the IP address of any of the rsync pods change in order to update rsync rac messaging addresses.

Documentation still work in progress
# rqlite-recover
