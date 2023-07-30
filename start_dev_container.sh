#!/bin/bash

docker run -d \
  --name=jellyfin \
  -e PUID=1000 \
  -e PGID=1000 \
  -e TZ=Etc/UTC \
  -e JELLYFIN_PublishedServerUrl=192.168.0.5 `#optional` \
  -p 8096:8096 \
  -v ./dev_container/config:/config \
  -v "$1":/data/media \
  --rm \
  lscr.io/linuxserver/jellyfin:latest
