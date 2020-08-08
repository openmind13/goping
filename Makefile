build:
	go build -o goping *.go && sudo ./goping


.DEFAULT_GOAL := build
