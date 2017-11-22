#!/bin/bash

echo "performing some demo load..."
for fn in echo-go echo-shell echo-python echo-node; do
  (for i in $(seq 1 1000); do
    echo "foobar" | btrfaasctl function invoke ${fn} >/dev/null
  done; echo "done with ${fn}"; ) &
done
echo ""
wait
