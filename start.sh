#!/bin/bash

function run() {
  go build . && (
    ./bla
  ) &
}

function stop() {
  pid=$(lsof -i -s TCP:LISTEN | awk 'NR==2 { print $2 }')
  if [ -n "${pid}" ]
  then
    kill ${pid}
  fi
}

trap stop EXIT SIGHUP SIGINT SIGQUIT SIGTERM

run

while inotifywait -qq -e modify *.html *.go
do
  stop
  date
  run
done
