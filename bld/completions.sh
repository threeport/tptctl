#!/bin/sh
set -ex
rm -rf completions
mkdir completions

for sh in bash zsh fish; do
	go run ./main.go completion "$sh" >"./completions/tptctl.$sh"
done
chmod +x ./completions/
