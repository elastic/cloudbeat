#!/usr/bin/env bash
set -euxo pipefail


non_blocking_wait() {
    PID=$1
    if [ ! -d "/proc/$PID" ]; then
        wait "$PID"
        CODE=$?
    else
        CODE=127
    fi
    return $CODE
}

get_logs () {
  kubectl logs --selector="catf=related" --all-containers=true --prefix -n kube-system --timestamps=true --since 10s
}

main () {
  just run-tests "$1" &
  PID=$!
  while true; do
    get_logs
    non_blocking_wait $PID
    CODE=$?
    if [ $CODE -ne 127 ]; then
        echo "PID $PID terminated with exit code $CODE"
        break
    fi
    sleep 10
  done
  echo "Tests Finished"
  cat test-logs/*.log | sed -n '/summary/,/===/p'
  kubectl delete po test-pod-v1 -n kube-system
  exit $CODE
}

main "$1"
