#!/usr/bin/env bash

echo "Checking that we are not using system ssh..."
ssh -V
[ $? -ne 255 ] && exit 1

echo "Exit code test"
ssh packer@fakessh exit 42
[ $? -ne 42 ] && exit 1

echo "Stderr check..."
packertest="$(2>&1 ssh packer@fakessh 'sh -c "printf packertest 1>&2"' )"
echo "Wanted packertest, Got $packertest"
[ "$packertest" != "packertest" ] && exit 1

echo "Sort check..."
printf '3\n1\n2' | ssh packer@fakessh 'sh -c "sort > sorted.txt"'
sorted="$(ssh packer@fakessh cat sorted.txt)"
printf "Wanted:\n1\n2\n3\nGot:\n$sorted\n"
[ "$sorted" = "1\n2\n3" ] && exit 1

exit 0
