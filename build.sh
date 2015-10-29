#!/usr/bin/env bash

GREEN='\e[1;32m'
RESET='\e[0m'

echo
echo -e $GREEN"Build information"
echo -e "--------------------------------------------------------------------------------"$RESET

echo "go env"
go env | sed "s/^/    /"
echo

echo "gitconfig"
git config -l | sed "s/^/    /"
echo

set -e

# Run 
go run main.go
