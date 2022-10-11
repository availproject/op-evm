#!/bin/sh

STAGED_GO_FILES=$(git diff --cached --name-only | grep ".go$")

if [[ "$STAGED_GO_FILES" = "" ]]; then
  exit 0
fi

GOLANGCI_LINT=$GOPATH/bin/golangci-lint

# Check for golangci-lint
if [[ ! -x "$GOLANGCI_LINT" ]]; then
  printf "\t\033[41mPlease install golangci-lint (go get -u github.com/golangci/golangci-lint/cmd/golangci-lint)"
  exit 1
fi

NORMAL=$(tput sgr0)
LIME_YELLOW=$(tput setaf 190)
printf "${LIME_YELLOW}Running golangci-lint on all staged *.go files...${NORMAL}\n"

RED=$(tput setaf 1)
GREEN=$(tput setaf 2)

golangci-lint run --out-format=github-actions --skip-dirs=third_party --timeout 5m
if [[ $? != 0 ]]; then
  printf "${RED}Linting failed! ${NORMAL}Please fix errors before committing.\n"
  exit 1
else
 printf "${GREEN}Linting passed! ${NORMAL}Continuing to commit.\n"
fi