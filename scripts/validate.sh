#!/bin/sh

set -eu

unformatted_files=$(gofmt -l .)

if [ -n "$unformatted_files" ]; then
	echo "The following Go files are not formatted:"
	printf '%s\n' "$unformatted_files"
	exit 1
fi

go vet ./...
go test ./...

echo "Validation passed."
